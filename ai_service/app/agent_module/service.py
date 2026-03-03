"""LangChain Agent Service — orchestrates tool calling, backend execution, and response generation.

Improvements:
  1. Complete Tool Calling integration (supports stateful multi-step intents like "add item to cart").
  2. Multi-turn reasoning loops enabled directly by DeepSeek function calling.
  3. Context-aware memory preservation via InMemoryChatMessageHistory.
"""

import json
import logging
import re
from difflib import SequenceMatcher
from typing import Any
from uuid import uuid4

from langchain_core.messages import SystemMessage, ToolMessage, AIMessage

from ..core.config import Settings
from ..langchain_module.history_store import ChatHistoryStore
from ..langchain_module.llm_factory import build_chat_llm, build_chat_llm_from_config
from ..langchain_module.text_utils import as_text, normalize_plain_text
from ..model_manager import ModelManager
from .backend_tools import BackendToolExecutor, ACTION_SPECS

logger = logging.getLogger(__name__)

# Human-readable order status mapping
ORDER_STATUS_MAP = {
    1: "待支付",
    2: "待发货",
    3: "运输中",
    4: "已完成",
    5: "已取消/已退款",
}

ADD_TO_CART_HINT_RE = re.compile(
    r"(加入|加进|加到|放入|放到|塞进|添加到|加购).{0,6}(购物车|购物袋|车里)|(购物车|购物袋|车里).{0,6}(加入|加进|加到|放入|放到|塞进|添加到|加购)"
)
DIRECT_PRODUCT_ID_RE = re.compile(r"(?:商品|product)\s*(?:id)?\s*[:：#]?\s*(\d+)", re.IGNORECASE)
QUANTITY_PREFIX_RE = re.compile(r"^(?P<count>\d+|两|俩|一|二|三|四|五)\s*(?:件|个|台|部|本|只|条|双)\s*")
ALNUM_TOKEN_RE = re.compile(r"[a-z0-9]+", re.IGNORECASE)
PRODUCT_NAME_PATTERNS = (
    re.compile(r"(?:帮我|请|麻烦|劳驾)?(?:把|将)?(?P<name>.+?)(?:加入|加进|加到|放入|放到|塞进|添加到|加购)(?:我的|当前)?(?:购物车|购物袋|车里)?$"),
    re.compile(r"(?:帮我|请|麻烦|劳驾)?(?:加入|加进|加到|放入|放到|塞进|添加到|加购)(?:我的|当前)?(?:购物车|购物袋|车里)?(?:里)?(?P<name>.+)$"),
)
LEADING_NOISE_RE = re.compile(r"^(?:把|将|这个|那个|这件|那件|这一件|那一件|这台|那台|这部|那部|这本|那本|这只|那只|一个|一件|一台|一部|一本|一只|一条|一双|最新的|全新的|那个叫|这款|那款)+")
TRAILING_NOISE_RE = re.compile(r"(?:加入购物车|加购物车|购物车|购物袋|车里|里|一下|吧|呀|呢|先|谢谢)+$")
DISPLAY_NAME_CLEAN_RE = re.compile(r"\s+")
QUANTITY_WORD_MAP = {"一": 1, "两": 2, "俩": 2, "二": 2, "三": 3, "四": 4, "五": 5}

class LangChainAgentService:
    def __init__(self, settings: Settings, history_store: ChatHistoryStore | None = None, model_manager: ModelManager | None = None) -> None:
        key = (settings.deepseek_api_key or "").strip()
        self._is_mock = not key or "replace_with_your_key" in key
        self._has_api_key = not self._is_mock
        self._settings = settings
        self._history_store = history_store or ChatHistoryStore()
        self._model_manager = model_manager
        self._executor = BackendToolExecutor(settings)
        
        if self._is_mock:
            self._llm = None
            self._tools = []
            self._system_prompt = None
            return

        # Instantiate LLM and Executor
        self._llm = build_chat_llm(settings)
        
        # Build Tool Manifests dynamically from ACTION_SPECS
        self._tools = []
        for spec in ACTION_SPECS:
            properties = {}
            for param in tuple(spec.path_params) + tuple(spec.query_params) + tuple(spec.body_params):
                if "id" in param or param in ("page", "page_size", "count", "status"):
                    param_type = "integer"
                elif param in ("is_random", "is_free_shipping", "is_negotiable", "is_home_delivery", "is_self_pickup"):
                    param_type = "boolean"
                elif param == "cart_ids":
                    param_type = "array"
                    properties[param] = {"type": "array", "items": {"type": "integer"}}
                    continue
                else:
                    param_type = "string"
                properties[param] = {"type": param_type}
                
            self._tools.append({
                "type": "function",
                "function": {
                    "name": spec.action.replace(".", "_"),
                    "description": f"{spec.description} {spec.param_hints}",
                    "parameters": {
                        "type": "object",
                        "properties": properties,
                        "required": []
                    }
                }
            })
            
        self._system_prompt = SystemMessage(content=(
            "你是闲趣二手交易平台的 AI 助手「闲趣小助手」。\n"
            "你能自动调用后端接口帮用户查商品、管订单、管购物车、管收藏等。\n\n"
            "## 核心工作流\n"
            "1. 当用户发出复合指令（如“把华为手机加入购物车”），你需要自动进行多步操作：先调用 `products_list` 拿到真实的商品ID，然后再根据ID调用 `cart_add` 加入购物车。\n"
            "2. 操作完成后，用自然、友好的中文告诉用户结果。如：“已为您把商品 [ID:x] 加入了购物车！”\n"
            "3. 回复必须是纯文本，不使用Markdown和代码块。\n"
            "4. 禁止执行“支付 (orders_pay / orders_confirm_pay)”相关的越权敏感操作，遇到这类请求请礼貌引导用户自行在页面点击支付。\n"
            "5. 所有列表数据，提取最重要的信息（商品名、价格、订单号），最多展示 3~5 条。\n"
            "6. 根据收到的返回 JSON，你需要将订单的状态码转成友好的文字告知：1=待支付，2=待发货，3=运输中，4=已完成，5=已取消/已退单\n"
        ))

    # -- public API ---------------------------------------------------------

    def chat(
        self,
        message: str,
        session_id: str | None = None,
        auth_token: str | None = None,
    ) -> tuple[str, str]:
        user_text = message.strip()
        if not user_text:
            raise ValueError("message is empty")

        use_session_id = session_id or uuid4().hex
        history = self._history_store.get_history(use_session_id)
        history.add_user_message(user_text)

        shortcut_answer = self._try_handle_add_to_cart_shortcut(user_text, auth_token)
        if shortcut_answer is not None:
            final_answer = normalize_plain_text(shortcut_answer)
            history.add_message(AIMessage(content=final_answer))
            return use_session_id, final_answer

        if self._is_mock and not self._model_manager:
            # Simple mock logic for Agent
            final_answer = f"我是闲趣小助手（演示模式）。由于 API Key 未正确设置，我目前无法为您查实真实数据或执行指令。您的请求是：{user_text}。"
            history.add_message(AIMessage(content=final_answer))
            return use_session_id, final_answer

        # Try dynamic model if model_manager is available
        active_llm = self._llm
        if self._model_manager:
            dynamic_llm = self._try_get_dynamic_llm()
            if dynamic_llm:
                active_llm = dynamic_llm

        if active_llm is None or self._executor is None:
             final_answer = "AI 代理服务未初始化，请检查配置。"
             history.add_message(AIMessage(content=final_answer))
             return use_session_id, final_answer
        
        # Feed the system prompt + whole history to LLM
        messages = [self._system_prompt] + history.messages # type: ignore
        llm_with_tools = active_llm.bind_tools(self._tools)
        
        # Max steps to prevent infinite tool loops
        final_answer = ""
        persist_final_answer = False
        for _ in range(8):
            response = llm_with_tools.invoke(messages)
            
            # Save the AI's partial/tool-calling thought process
            messages.append(response)
            history.add_message(response)
            
            if not getattr(response, "tool_calls", None):
                final_answer = as_text(response.content)
                break
                
            # Execute multiple parallel tool calls
            for tc in response.tool_calls:
                action_name = tc["name"].replace("_", ".")
                params = tc.get("args", {})
                logger.info("[Agent] Calling Tool: %s(%s)", action_name, params)
                
                try:
                    result = self._executor.execute_action(action_name, params, auth_token)
                    content = self._summarize_result(result)
                    logger.info("[Agent] Result -> %s", content[:200])
                except Exception as e:
                    content = json.dumps({"ok": False, "error": str(e)}, ensure_ascii=False)
                    logger.warning("[Agent] Error -> %s", content)
                    
                # Append tool execution result back to messages & history
                tm = ToolMessage(content=content[:1500], tool_call_id=tc["id"])
                messages.append(tm)
                history.add_message(tm)
                
        if not final_answer:
            final_answer = "我执行的操作有些复杂，似乎遇到了一点网络波动，请稍后再试一次。"
            persist_final_answer = True
            
        final_answer = normalize_plain_text(final_answer)
        if persist_final_answer:
            history.add_message(AIMessage(content=final_answer))
        return use_session_id, final_answer

    def clear_session(self, session_id: str) -> bool:
        return self._history_store.clear_session(session_id)

    def _try_get_dynamic_llm(self):
        """Try to get a ChatOpenAI instance from dynamically configured models."""
        if not self._model_manager:
            return None
        models = self._model_manager.fetch_active_models()
        for model_cfg in models:
            secret = self._model_manager.fetch_model_secret(model_cfg.id)
            if not secret or not secret.api_key:
                continue
            try:
                llm = build_chat_llm_from_config(
                    secret,
                    temperature=self._settings.deepseek_temperature,
                    timeout_seconds=self._settings.deepseek_timeout,
                )
                logger.info(f"Agent using dynamic model: {secret.provider}/{secret.model_name}")
                return llm
            except Exception as e:
                logger.warning(f"Failed to build LLM for model {model_cfg.id}: {e}")
                continue
        return None

    # -- private helpers ---------------------------------------------------

    def _try_handle_add_to_cart_shortcut(self, user_text: str, auth_token: str | None) -> str | None:
        if self._executor is None or not ADD_TO_CART_HINT_RE.search(user_text):
            return None

        parsed = self._parse_add_to_cart_request(user_text)
        if not parsed:
            return "我听出来您想加购物车，但没认清商品名。请直接说商品全名，比如“把 iPhone 15 Pro Max 加入购物车”。"

        if not auth_token:
            return "请先登录，再让我帮您把商品加入购物车。"

        direct_product_id = parsed.get("product_id")
        if isinstance(direct_product_id, int) and direct_product_id > 0:
            product_detail = self._executor.execute_action("products.detail", {"id": direct_product_id}, auth_token)
            if not product_detail.get("ok"):
                return f"没找到商品 ID {direct_product_id}，请检查后再试。"
            product_data = self._unwrap_product_detail(product_detail)
            if not product_data:
                return f"没找到商品 ID {direct_product_id}，请检查后再试。"
            return self._add_product_to_cart(product_data, int(parsed.get("count", 1) or 1), auth_token)

        product_name = str(parsed.get("product_name") or "").strip()
        if not product_name:
            return "我听出来您想加购物车，但没认清商品名。请再说完整一点。"

        product = self._find_best_matching_product(product_name, auth_token)
        if not product:
            return f"未找到“{product_name}”，请检查名称或试试更短的关键词。"

        return self._add_product_to_cart(product, int(parsed.get("count", 1) or 1), auth_token, requested_name=product_name)

    def _parse_add_to_cart_request(self, user_text: str) -> dict[str, Any] | None:
        text = " ".join(user_text.strip().split())
        if not text:
            return None

        direct_id_match = DIRECT_PRODUCT_ID_RE.search(text)
        if direct_id_match:
            return {"product_id": int(direct_id_match.group(1)), "count": 1}

        product_name = ""
        for pattern in PRODUCT_NAME_PATTERNS:
            match = pattern.search(text)
            if match:
                product_name = match.group("name")
                break

        if not product_name:
            quoted_match = re.search(r"[“\"'「『](.+?)[”\"'」』]", text)
            if quoted_match:
                product_name = quoted_match.group(1)

        if not product_name:
            return None

        product_name = product_name.strip()
        product_name = LEADING_NOISE_RE.sub("", product_name)
        product_name = TRAILING_NOISE_RE.sub("", product_name)
        product_name = product_name.strip("：:，,。.!！？?、 ")

        count = 1
        quantity_match = QUANTITY_PREFIX_RE.match(product_name)
        if quantity_match:
            raw_count = quantity_match.group("count")
            count = int(raw_count) if raw_count.isdigit() else QUANTITY_WORD_MAP.get(raw_count, 1)
            product_name = product_name[quantity_match.end():].strip()

        product_name = product_name.strip("“”\"'‘’()（）[]【】")
        if not product_name:
            return None

        return {"product_name": product_name, "count": max(1, min(count, 9))}

    def _find_best_matching_product(self, product_name: str, auth_token: str | None) -> dict[str, Any] | None:
        candidates: list[dict[str, Any]] = []
        seen_ids: set[int] = set()
        search_queries = self._build_product_search_queries(product_name)

        for query in search_queries:
            result = self._executor.execute_action(
                "products.list",
                {"search": query, "page": 1, "page_size": 20},
                auth_token,
            )
            if not result.get("ok"):
                continue
            for item in self._extract_products(result):
                product_id = int(item.get("id") or 0)
                if product_id and product_id not in seen_ids:
                    candidates.append(item)
                    seen_ids.add(product_id)

        if not candidates:
            return None

        best_product, best_score, second_score = self._score_product_candidates(product_name, candidates)
        if best_product is None or best_score < 0.58:
            return None
        if second_score >= 0.72 and (best_score - second_score) < 0.08:
            return None
        return best_product

    def _build_product_search_queries(self, product_name: str) -> list[str]:
        queries: list[str] = []
        cleaned = DISPLAY_NAME_CLEAN_RE.sub(" ", product_name.strip())
        compact = re.sub(r"[\s\-_/]+", "", cleaned)
        alpha_tokens = ALNUM_TOKEN_RE.findall(cleaned.lower())

        for value in (cleaned, compact):
            if value and value not in queries:
                queries.append(value)

        if len(alpha_tokens) >= 2:
            joined = " ".join(alpha_tokens)
            if joined not in queries:
                queries.append(joined)
            squashed = "".join(alpha_tokens)
            if squashed not in queries:
                queries.append(squashed)

        if alpha_tokens and alpha_tokens[0] not in queries:
            queries.append(alpha_tokens[0])

        return queries[:5]

    def _score_product_candidates(
        self,
        requested_name: str,
        candidates: list[dict[str, Any]],
    ) -> tuple[dict[str, Any] | None, float, float]:
        requested_norm = self._normalize_product_name(requested_name)
        scored: list[tuple[float, dict[str, Any]]] = []

        for item in candidates:
            product_name = str(item.get("name") or "").strip()
            if not product_name:
                continue

            score = 0.0
            product_norm = self._normalize_product_name(product_name)
            if requested_norm == product_norm:
                score += 1.4
            if requested_norm and requested_norm in product_norm:
                score += 1.0
            if product_norm and product_norm in requested_norm:
                score += 0.65
            score += SequenceMatcher(None, requested_norm, product_norm).ratio() * 0.6

            description_norm = self._normalize_product_name(str(item.get("description") or ""))
            if requested_norm and requested_norm in description_norm:
                score += 0.12

            if int(item.get("status") or 0) != 1:
                score -= 0.3
            if int(item.get("count") or 1) < 1:
                score -= 0.2

            scored.append((score, item))

        if not scored:
            return None, 0.0, 0.0

        scored.sort(key=lambda entry: entry[0], reverse=True)
        best_score, best_product = scored[0]
        second_score = scored[1][0] if len(scored) > 1 else 0.0
        return best_product, best_score, second_score

    def _normalize_product_name(self, value: str) -> str:
        text = str(value or "").strip().lower()
        if not text:
            return ""
        text = text.replace("＋", "+")
        text = re.sub(r"[“”\"'‘’`·•]+", "", text)
        text = re.sub(r"[^\w\u4e00-\u9fff]+", "", text, flags=re.UNICODE)
        return text

    def _extract_products(self, result: dict[str, Any]) -> list[dict[str, Any]]:
        data = result.get("data")
        if not isinstance(data, dict):
            return []
        items = data.get("list")
        if not isinstance(items, list):
            return []
        return [item for item in items if isinstance(item, dict)]

    def _unwrap_product_detail(self, result: dict[str, Any]) -> dict[str, Any] | None:
        data = result.get("data")
        if isinstance(data, dict) and isinstance(data.get("data"), dict):
            return data.get("data")
        if isinstance(data, dict):
            return data
        return None

    def _add_product_to_cart(
        self,
        product: dict[str, Any],
        count: int,
        auth_token: str | None,
        requested_name: str | None = None,
    ) -> str:
        product_id = int(product.get("id") or 0)
        display_name = str(product.get("name") or requested_name or "该商品").strip()
        if not product_id:
            return f"没能识别出“{requested_name or display_name}”的商品编号，请稍后再试。"

        add_result = self._executor.execute_action(
            "cart.add",
            {"product_id": product_id, "count": max(1, count)},
            auth_token,
        )
        if not add_result.get("ok"):
            return add_result.get("message") or f"“{display_name}”加入购物车失败，请稍后重试。"

        api_message = str(add_result.get("data", {}).get("message") or "").strip()
        if "数量已更新" in api_message:
            return f"已成功将“{display_name}”加入购物车，购物车里的数量也同步更新了。"
        return f"已成功将“{display_name}”加入购物车。"

    def _summarize_result(self, result: dict[str, Any]) -> str:
        """Summarize a single action result, enriching with human-readable data."""
        if not result.get("ok"):
            return json.dumps(result, ensure_ascii=False, separators=(",", ":"))

        data = result.get("data")
        action = result.get("action", "")

        # For product list, extract key fields
        if action == "products.list" and isinstance(data, dict):
            items = data.get("list", [])
            total = data.get("total", len(items))
            if items:
                summaries = []
                for item in items[:8]:  # Cap at 8 items for context window
                    summaries.append({
                        "id": item.get("id", ""),
                        "name": item.get("name", "未知"),
                        "price": item.get("price", 0),
                        "category": item.get("category", ""),
                        "status": item.get("status", 0),
                        "count": item.get("count", 1),
                    })
                return json.dumps(
                    {"ok": True, "summary": f"找到{total}件商品", "items": summaries},
                    ensure_ascii=False,
                    separators=(",", ":"),
                )
            return json.dumps({"ok": True, "summary": "没有找到任何商品", "items": []}, ensure_ascii=False, separators=(",", ":"))

        if action == "products.detail" and isinstance(data, dict):
            d = data.get("data", data)
            return json.dumps(d, ensure_ascii=False, separators=(",", ":"))[:600]

        # For order list
        if action == "orders.list" and isinstance(data, dict):
            orders = data.get("data", [])
            if not orders:
                return "{\"ok\": true, \"summary\": \"没有订单记录\"}"
            summaries = []
            for o in orders[:8]:
                oid = o.get("id", "")
                order_no = o.get("order_no", "")
                price = o.get("price", 0)
                status = ORDER_STATUS_MAP.get(o.get("status", 0), "未知")
                product = o.get("product", {})
                pname = product.get("name", "未知商品")
                summaries.append(f"[订单ID:{oid}] {order_no} {pname} ¥{price:.2f} ({status})")
            return f"{{\"ok\": true, \"items\": {json.dumps(summaries, ensure_ascii=False)}}}"

        # For cart list
        if action == "cart.list" and isinstance(data, dict):
            items = data.get("data", [])
            if not items:
                return "{\"ok\": true, \"summary\": \"购物车为空\"}"
            summaries = []
            for item in items[:8]:
                cid = item.get("id", "")
                count = item.get("count", 1)
                product = item.get("product", {})
                pid = product.get("id", "")
                pname = product.get("name", "未知")
                price = product.get("price", 0)
                summaries.append(f"[购物车ID:{cid} / 商品ID:{pid}] {pname} ¥{price:.2f} x{count}")
            return f"{{\"ok\": true, \"items\": {json.dumps(summaries, ensure_ascii=False)}}}"

        # For favorites (user.data with type=favorites)
        if action == "user.data" and isinstance(data, dict):
            items = data.get("data", [])
            if not items:
                return "{\"ok\": true, \"summary\": \"数据为空\"}"
            if isinstance(items, list) and items and "product" in items[0]:
                summaries = []
                for fav in items[:8]:
                    product = fav.get("product", {})
                    pid = product.get("id", "")
                    pname = product.get("name", "未知")
                    price = product.get("price", 0)
                    summaries.append(f"[商品ID:{pid}] {pname} ¥{price:.2f}")
                return f"{{\"ok\": true, \"items\": {json.dumps(summaries, ensure_ascii=False)}}}"
            return json.dumps(data, ensure_ascii=False, separators=(",", ":"))[:600]

        # Default: dump the result JSON string truncated
        return json.dumps({"ok": True, "action": action, "data": data}, ensure_ascii=False, separators=(",", ":"))[:600]

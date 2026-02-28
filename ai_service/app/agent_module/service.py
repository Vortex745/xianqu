"""LangChain Agent Service — orchestrates tool calling, backend execution, and response generation.

Improvements:
  1. Complete Tool Calling integration (supports stateful multi-step intents like "add item to cart").
  2. Multi-turn reasoning loops enabled directly by DeepSeek function calling.
  3. Context-aware memory preservation via InMemoryChatMessageHistory.
"""

import json
from typing import Any
from uuid import uuid4

from langchain_core.messages import HumanMessage, SystemMessage, ToolMessage, AIMessage

from ..core.config import Settings
from ..langchain_module.history_store import ChatHistoryStore
from ..langchain_module.llm_factory import build_chat_llm, build_chat_llm_from_config
from ..langchain_module.text_utils import as_text, normalize_plain_text
from ..model_manager import ModelManager
from .backend_tools import BackendToolExecutor, ACTION_SPECS

import logging

logger = logging.getLogger(__name__)

# Human-readable order status mapping
ORDER_STATUS_MAP = {
    1: "待支付",
    2: "待发货",
    3: "运输中",
    4: "已完成",
    5: "已取消/已退款",
}

class LangChainAgentService:
    def __init__(self, settings: Settings, history_store: ChatHistoryStore | None = None, model_manager: ModelManager | None = None) -> None:
        key = (settings.deepseek_api_key or "").strip()
        self._is_mock = not key or "replace_with_your_key" in key
        self._has_api_key = not self._is_mock
        self._settings = settings
        self._history_store = history_store or ChatHistoryStore()
        self._model_manager = model_manager
        
        if self._is_mock:
            self._llm = None
            self._executor = None
            self._tools = []
            self._system_prompt = None
            return

        # Instantiate LLM and Executor
        self._llm = build_chat_llm(settings)
        self._executor = BackendToolExecutor(settings)
        
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

        if self._is_mock and not self._model_manager:
            # Simple mock logic for Agent
            return use_session_id, f"我是闲趣小助手（演示模式）。由于 API Key 未正确设置，我目前无法为您查实真实数据或执行指令。 您的请求是：{user_text}。"

        # Try dynamic model if model_manager is available
        active_llm = self._llm
        if self._model_manager:
            dynamic_llm = self._try_get_dynamic_llm()
            if dynamic_llm:
                active_llm = dynamic_llm

        if active_llm is None or self._executor is None:
             return use_session_id, "AI 代理服务未初始化，请检查配置。"

        history = self._history_store.get_history(use_session_id)
        
        # Append the new human message to our persistent history
        history.add_user_message(user_text)
        
        # Feed the system prompt + whole history to LLM
        messages = [self._system_prompt] + history.messages # type: ignore
        llm_with_tools = active_llm.bind_tools(self._tools)
        
        # Max steps to prevent infinite tool loops
        final_answer = ""
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
                print(f"[Agent] Calling Tool: {action_name}({params})")
                
                try:
                    result = self._executor.execute_action(action_name, params, auth_token)
                    content = self._summarize_result(result)
                    print(f"[Agent] Result -> {content[:200]}")
                except Exception as e:
                    content = json.dumps({"ok": False, "error": str(e)}, ensure_ascii=False)
                    print(f"[Agent] Error -> {content}")
                    
                # Append tool execution result back to messages & history
                tm = ToolMessage(content=content[:1500], tool_call_id=tc["id"])
                messages.append(tm)
                history.add_message(tm)
                
        if not final_answer:
            final_answer = "我执行的操作有些复杂，似乎遇到了一点网络波动，请稍后再试一次。"
            
        final_answer = normalize_plain_text(final_answer)
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
                    name = item.get("name", "未知")
                    price = item.get("price", 0)
                    category = item.get("category", "")
                    status = "在售" if item.get("status") == 1 else "已售"
                    pid = item.get("id", "")
                    summaries.append(f"[ID:{pid}] {name} ¥{price:.2f} {category} ({status})")
                return f"OK"  # For internal tool chaining, returning full JSON is often better, but summarizing saves tokens. 
                # Let's return the structured summary back
                return f"{{\"ok\": true, \"summary\": \"找到{total}件商品\", \"items\": {json.dumps(summaries, ensure_ascii=False)}}}"
            return "{\"ok\": true, \"summary\": \"没有找到任何商品\"}"

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

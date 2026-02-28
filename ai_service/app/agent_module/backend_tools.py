"""Backend tool executor — white-listed HTTP actions for the AI Agent.

Covers **all** non-payment backend APIs:
  - Product: list, detail, search, categories, create, update
  - Order: list, create, batch_create, confirm_receive, refund (cancel)
  - Cart: list, add, delete
  - Favorites: check, add, remove
  - User: profile, data, public info, password change
  - Chat: contacts, messages

Excluded (payment): orders.pay, orders.confirm_pay
"""

from dataclasses import dataclass, field
from typing import Any

import httpx

from ..core.config import Settings


@dataclass(frozen=True)
class ActionSpec:
    """Specification for a single backend action."""

    action: str
    method: str
    path_template: str
    requires_auth: bool
    description: str
    path_params: tuple[str, ...] = ()
    query_params: tuple[str, ...] = ()
    body_params: tuple[str, ...] = ()
    # Extra hint shown to LLM so it knows what params are valid
    param_hints: str = ""


# ---------------------------------------------------------------------------
# Complete action catalogue (no payment)
# ---------------------------------------------------------------------------

ACTION_SPECS: tuple[ActionSpec, ...] = (
    # ── Product ────────────────────────────────────────────────────────────
    ActionSpec(
        action="products.list",
        method="GET",
        path_template="/products",
        requires_auth=False,
        description="查询商品列表、搜索商品、按分类/区域筛选、随机推荐。",
        query_params=("page", "page_size", "category", "search", "is_random", "area"),
        param_hints="search=关键词  category=数码|书籍|生活|服饰|运动|美妆|乐器|手游|兼职|其他  area=区域名  is_random=true 随机推荐  page/page_size 分页",
    ),
    ActionSpec(
        action="products.detail",
        method="GET",
        path_template="/products/{id}",
        requires_auth=False,
        description="查看指定商品详情（价格、描述、卖家、收藏数、浏览量等）。",
        path_params=("id",),
        param_hints="id=商品ID(数字)",
    ),
    ActionSpec(
        action="products.categories",
        method="GET",
        path_template="/categories",
        requires_auth=False,
        description="获取所有商品分类列表。",
    ),
    ActionSpec(
        action="products.create",
        method="POST",
        path_template="/products",
        requires_auth=True,
        description="发布新商品。",
        body_params=(
            "name", "description", "price", "count", "image", "category",
            "area", "is_free_shipping", "is_negotiable", "is_home_delivery", "is_self_pickup",
        ),
        param_hints="name=商品名(必填) price=价格(必填) category=分类(必填) 其余可选",
    ),
    ActionSpec(
        action="products.update",
        method="PUT",
        path_template="/products/{id}",
        requires_auth=True,
        description="更新已发布的商品信息。",
        path_params=("id",),
        body_params=(
            "name", "description", "price", "count", "image", "category",
            "area", "status", "is_free_shipping", "is_negotiable", "is_home_delivery", "is_self_pickup",
        ),
        param_hints="id=商品ID  只传需要改的字段",
    ),
    # ── Order ──────────────────────────────────────────────────────────────
    ActionSpec(
        action="orders.list",
        method="GET",
        path_template="/orders",
        requires_auth=True,
        description="查询我的订单列表。role=buyer 买家订单，role=seller 卖家订单。",
        query_params=("role",),
        param_hints="role=buyer(默认)|seller。返回包含商品名、价格、订单状态等信息。状态: 1=待支付 2=待发货 3=运输中 4=已完成 5=已取消",
    ),
    ActionSpec(
        action="orders.create",
        method="POST",
        path_template="/orders",
        requires_auth=True,
        description="创建单商品订单（直接购买）。",
        body_params=("product_id",),
        param_hints="product_id=商品ID(必填)",
    ),
    ActionSpec(
        action="orders.batch_create",
        method="POST",
        path_template="/orders/batch",
        requires_auth=True,
        description="购物车批量结算下单。",
        body_params=("cart_ids",),
        param_hints="cart_ids=[购物车ID列表](必填)",
    ),
    ActionSpec(
        action="orders.confirm_receive",
        method="PUT",
        path_template="/orders/{id}/confirm",
        requires_auth=True,
        description="确认收货，将订单标记为已完成。",
        path_params=("id",),
        param_hints="id=订单ID",
    ),
    ActionSpec(
        action="orders.refund",
        method="PUT",
        path_template="/orders/{id}/refund",
        requires_auth=True,
        description="取消订单 / 申请退单。待支付(1)待发货(2)运输中(3)的订单可取消。",
        path_params=("id",),
        param_hints="id=订单ID",
    ),
    # ── Cart ───────────────────────────────────────────────────────────────
    ActionSpec(
        action="cart.list",
        method="GET",
        path_template="/cart",
        requires_auth=True,
        description="查看我的购物车列表（含商品名称、价格、数量等）。",
    ),
    ActionSpec(
        action="cart.add",
        method="POST",
        path_template="/cart/add",
        requires_auth=True,
        description="添加商品到购物车。",
        body_params=("product_id", "count"),
        param_hints="product_id=商品ID(必填) count=数量(默认1)",
    ),
    ActionSpec(
        action="cart.delete",
        method="DELETE",
        path_template="/cart/{id}",
        requires_auth=True,
        description="删除购物车中的指定条目。",
        path_params=("id",),
        param_hints="id=购物车条目ID",
    ),
    # ── Favorites ──────────────────────────────────────────────────────────
    ActionSpec(
        action="favorites.check",
        method="GET",
        path_template="/favorites/check",
        requires_auth=True,
        description="检查某商品是否已被我收藏。",
        query_params=("product_id",),
        param_hints="product_id=商品ID",
    ),
    ActionSpec(
        action="favorites.add",
        method="POST",
        path_template="/favorites/add",
        requires_auth=True,
        description="收藏指定商品。",
        body_params=("product_id",),
        param_hints="product_id=商品ID",
    ),
    ActionSpec(
        action="favorites.remove",
        method="POST",
        path_template="/favorites/remove",
        requires_auth=True,
        description="取消收藏指定商品。",
        body_params=("product_id",),
        param_hints="product_id=商品ID",
    ),
    # ── User ───────────────────────────────────────────────────────────────
    ActionSpec(
        action="user.data",
        method="GET",
        path_template="/user/data",
        requires_auth=True,
        description="查询我的发布(type=products)或收藏(type=favorites)数据。",
        query_params=("type",),
        param_hints="type=products|favorites",
    ),
    ActionSpec(
        action="user.profile.update",
        method="PUT",
        path_template="/user/profile",
        requires_auth=True,
        description="更新个人资料（昵称、头像、手机号）。",
        body_params=("nickname", "avatar", "phone"),
        param_hints="nickname=新昵称 avatar=头像URL phone=手机号 只传需要改的字段",
    ),
    ActionSpec(
        action="user.password.change",
        method="PUT",
        path_template="/user/password",
        requires_auth=True,
        description="修改账号密码。",
        body_params=("old_password", "new_password"),
        param_hints="old_password=旧密码(必填) new_password=新密码(必填)",
    ),
    ActionSpec(
        action="user.public",
        method="GET",
        path_template="/users/{id}",
        requires_auth=True,
        description="查看指定用户的公开信息（昵称、头像）。",
        path_params=("id",),
        param_hints="id=用户ID",
    ),
    # ── Chat ───────────────────────────────────────────────────────────────
    ActionSpec(
        action="chat.contacts",
        method="GET",
        path_template="/chat/contacts",
        requires_auth=True,
        description="查看我的聊天联系人列表。",
    ),
    ActionSpec(
        action="chat.messages",
        method="GET",
        path_template="/chat/messages",
        requires_auth=True,
        description="查看与指定联系人的聊天记录。",
        query_params=("target_id",),
        param_hints="target_id=对方用户ID",
    ),
)

ACTION_SPEC_MAP: dict[str, ActionSpec] = {spec.action: spec for spec in ACTION_SPECS}


# ---------------------------------------------------------------------------
# Prompt helpers
# ---------------------------------------------------------------------------

def action_catalog_for_prompt() -> str:
    """Build a structured action catalogue string for the LLM planner."""
    lines: list[str] = []
    for spec in ACTION_SPECS:
        auth = "🔒需登录" if spec.requires_auth else "🌐公开"
        hint = f"  参数: {spec.param_hints}" if spec.param_hints else ""
        lines.append(f"- {spec.action} ({auth}): {spec.description}{hint}")
    return "\n".join(lines)


def supported_actions() -> set[str]:
    s = set(ACTION_SPEC_MAP.keys())
    s.add("chat_only")           # virtual action for pure conversation
    s.add("multi_step")          # virtual action for multi-step tasks (e.g. compare)
    return s


# ---------------------------------------------------------------------------
# Auth token helpers
# ---------------------------------------------------------------------------

def normalize_auth_token(raw_token: str | None) -> str:
    value = str(raw_token or "").strip()
    if not value:
        return ""
    if value.lower().startswith("bearer "):
        return value
    return f"Bearer {value}"


# ---------------------------------------------------------------------------
# Param extraction helpers
# ---------------------------------------------------------------------------

def _resolve_param(params: dict[str, Any], key: str) -> Any:
    """Look up *key* from flat params or nested containers."""
    if key in params:
        return params[key]
    for container_key in ("params", "query", "body", "payload"):
        container = params.get(container_key)
        if isinstance(container, dict) and key in container:
            return container[key]
    return None


def _extract_allowed_params(params: dict[str, Any], keys: tuple[str, ...]) -> dict[str, Any]:
    payload: dict[str, Any] = {}
    for key in keys:
        value = _resolve_param(params, key)
        if value is not None:
            payload[key] = value
    return payload


# ---------------------------------------------------------------------------
# Executor
# ---------------------------------------------------------------------------

class BackendToolExecutor:
    """Executes white-listed HTTP requests against the Go backend."""

    def __init__(self, settings: Settings) -> None:
        self._base_url = settings.backend_api_base_url.rstrip("/")
        self._timeout = settings.backend_timeout

    # -- public API ---------------------------------------------------------

    def execute_action(
        self,
        action: str,
        params: dict[str, Any] | None,
        auth_token: str | None = None,
    ) -> dict[str, Any]:
        spec = ACTION_SPEC_MAP.get(action)
        if not spec:
            return {
                "ok": False,
                "action": action,
                "error": "unsupported_action",
                "message": "当前指令不在可调用白名单内。",
            }

        payload = params if isinstance(params, dict) else {}
        token = normalize_auth_token(auth_token)
        if spec.requires_auth and not token:
            return {
                "ok": False,
                "action": action,
                "error": "missing_auth",
                "message": "该操作需要登录后执行。请先登录再操作。",
            }

        # Resolve path params
        try:
            path_values = _extract_allowed_params(payload, spec.path_params)
            for key in spec.path_params:
                if key not in path_values:
                    return {
                        "ok": False,
                        "action": action,
                        "error": "missing_path_param",
                        "message": f"缺少必要参数：{key}",
                    }
            path = spec.path_template.format(**path_values)
        except Exception:
            return {
                "ok": False,
                "action": action,
                "error": "invalid_path_param",
                "message": "路径参数格式无效。",
            }

        query = _extract_allowed_params(payload, spec.query_params)
        body = _extract_allowed_params(payload, spec.body_params)

        # Allow explicit nested payload override
        payload_body = payload.get("payload")
        if isinstance(payload_body, dict):
            for key in spec.body_params:
                if key in payload_body:
                    body[key] = payload_body[key]

        request_json = body if spec.method in {"POST", "PUT", "PATCH"} else None

        return self._request(
            method=spec.method,
            path=path,
            token=token,
            query=query if query else None,
            body=request_json,
            action=action,
        )

    def execute_multi_step(
        self,
        steps: list[dict[str, Any]],
        auth_token: str | None = None,
    ) -> list[dict[str, Any]]:
        """Execute a sequence of actions and return all results."""
        results: list[dict[str, Any]] = []
        for step in steps:
            action = step.get("action", "")
            params = step.get("params", {})
            result = self.execute_action(action, params, auth_token)
            results.append(result)
        return results

    # -- private -----------------------------------------------------------

    def _request(
        self,
        method: str,
        path: str,
        token: str,
        query: dict[str, Any] | None,
        body: dict[str, Any] | None,
        action: str,
    ) -> dict[str, Any]:
        headers: dict[str, str] = {"Accept": "application/json"}
        if token:
            headers["Authorization"] = token

        url = f"{self._base_url}{path}"
        try:
            with httpx.Client(timeout=self._timeout, trust_env=False) as client:
                response = client.request(
                    method=method,
                    url=url,
                    headers=headers,
                    params=query,
                    json=body,
                )
        except httpx.RequestError as exc:
            return {
                "ok": False,
                "action": action,
                "error": "network_error",
                "message": f"后端服务请求失败：{exc}",
            }

        parsed: Any = None
        raw_text = response.text.strip()
        if raw_text:
            try:
                parsed = response.json()
            except ValueError:
                parsed = raw_text

        if response.status_code >= 400:
            return {
                "ok": False,
                "action": action,
                "status_code": response.status_code,
                "error": "api_error",
                "message": _extract_error_message(parsed, raw_text, response.status_code),
                "data": parsed,
            }

        return {
            "ok": True,
            "action": action,
            "status_code": response.status_code,
            "data": parsed,
        }


def _extract_error_message(parsed: Any, fallback_text: str, status_code: int) -> str:
    if isinstance(parsed, dict):
        error_text = parsed.get("error") or parsed.get("message")
        if isinstance(error_text, str) and error_text.strip():
            return error_text.strip()
    if fallback_text:
        return fallback_text
    return f"接口返回状态异常：{status_code}"

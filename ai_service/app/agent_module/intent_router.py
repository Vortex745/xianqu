"""Intent router — maps natural-language user input to an action plan.

Improvements over the original version:
  1. Structured system prompt with examples for few-shot learning
  2. Escaped JSON brackets to avoid LangChain templating errors
  3. Support for multi_step actions (e.g. price comparison)
  4. Better intent recognition for Chinese shopping scenarios
"""

import json
import re
from dataclasses import dataclass
from typing import Any

from langchain_core.prompts import ChatPromptTemplate

from ..langchain_module.text_utils import as_text
from .backend_tools import action_catalog_for_prompt, supported_actions


@dataclass(frozen=True)
class AgentPlan:
    intent: str
    action: str
    params: dict[str, Any]
    # For multi-step tasks like comparison
    steps: list[dict[str, Any]] | None = None


_SYSTEM_PROMPT = """\
你是闲趣二手交易平台的 AI Agent 意图路由器。
你的唯一职责：把用户自然语言输入解析成一个 JSON 执行计划。

## 核心规则
1. 仅输出 JSON，不输出任何解释文字
2. 禁止调用支付（pay / confirm_pay）相关动作
3. 如果用户只是闲聊或询问平台使用方法，action 设为 chat_only
4. 当用户要比价时，使用 multi_step action，steps 里包含多个 products.list 调用
5. 取消订单用 orders.refund（不是 delete）
6. 搜索商品用 products.list + search 参数

## 订单状态说明
1=待支付  2=待发货  3=运输中  4=已完成  5=已取消/已退款

## 分类可选值
数码、书籍、生活、服饰、运动、美妆、乐器、手游、兼职、其他

## 输出格式
单步动作：
{{"intent":"一句话意图","action":"action名","params":{{"键":"值"}}}}

多步动作（如比价）：
{{"intent":"比价意图","action":"multi_step","params":{{}},"steps":[{{"action":"products.list","params":{{"search":"关键词1"}}}},{{"action":"products.list","params":{{"search":"关键词2"}}}}]}}

## 示例

用户: 帮我搜一下耳机
{{"intent":"搜索耳机商品","action":"products.list","params":{{"search":"耳机"}}}}

用户: 查看商品 42 的详情
{{"intent":"查看商品详情","action":"products.detail","params":{{"id":42}}}}

用户: 帮我取消订单 10
{{"intent":"取消订单","action":"orders.refund","params":{{"id":10}}}}

用户: 帮我查一下我买的订单
{{"intent":"查看买家订单","action":"orders.list","params":{{"role":"buyer"}}}}

用户: 帮我查卖家订单
{{"intent":"查看卖家订单","action":"orders.list","params":{{"role":"seller"}}}}

用户: 帮我比较一下华为和苹果耳机的价格
{{"intent":"比较华为和苹果耳机价格","action":"multi_step","params":{{}},"steps":[{{"action":"products.list","params":{{"search":"华为耳机"}}}},{{"action":"products.list","params":{{"search":"苹果耳机"}}}}]}}

用户: 有什么运动类的商品
{{"intent":"查看运动类商品","action":"products.list","params":{{"category":"运动"}}}}

用户: 看看我的购物车
{{"intent":"查看购物车","action":"cart.list","params":{{}}}}

用户: 看看我收藏了什么
{{"intent":"查看我的收藏","action":"user.data","params":{{"type":"favorites"}}}}

用户: 帮我把商品 8 加入购物车
{{"intent":"加入购物车","action":"cart.add","params":{{"product_id":8,"count":1}}}}

用户: 帮我收藏商品 15
{{"intent":"收藏商品","action":"favorites.add","params":{{"product_id":15}}}}

用户: 确认收货订单 20
{{"intent":"确认收货","action":"orders.confirm_receive","params":{{"id":20}}}}

用户: 你好
{{"intent":"打招呼","action":"chat_only","params":{{}}}}

用户: 帮我改昵称为小明
{{"intent":"修改昵称","action":"user.profile.update","params":{{"nickname":"小明"}}}}
"""


class IntentRouter:
    def __init__(self, llm: Any) -> None:
        self._llm = llm
        self._allowed_actions = supported_actions()
        self._planner = ChatPromptTemplate.from_messages(
            [
                ("system", _SYSTEM_PROMPT),
                (
                    "human",
                    "可用 action 列表：\n"
                    "{action_catalog}\n\n"
                    "用户输入：{input}",
                ),
            ]
        )

    def plan(self, user_message: str) -> AgentPlan:
        text = user_message.strip()
        if not text:
            return AgentPlan(intent="空输入", action="chat_only", params={})

        response = (self._planner | self._llm).invoke(
            {
                "input": text,
                "action_catalog": action_catalog_for_prompt(),
            }
        )
        payload = self._parse_plan(as_text(response.content))
        action = payload.get("action") if isinstance(payload.get("action"), str) else "chat_only"
        if action not in self._allowed_actions:
            action = "chat_only"

        intent = payload.get("intent") if isinstance(payload.get("intent"), str) else "通用咨询"
        params = payload.get("params") if isinstance(payload.get("params"), dict) else {}
        steps = payload.get("steps") if isinstance(payload.get("steps"), list) else None

        return AgentPlan(intent=intent, action=action, params=params, steps=steps)

    def _parse_plan(self, raw: str) -> dict[str, Any]:
        text = raw.strip()
        if not text:
            return {}

        # Remove possible fenced block wrappers
        text = re.sub(r"^```(?:json)?\s*", "", text, flags=re.IGNORECASE)
        text = re.sub(r"\s*```$", "", text)

        # Try direct parse
        try:
            data = json.loads(text)
            if isinstance(data, dict):
                return data
        except json.JSONDecodeError:
            pass

        # Try extracting first JSON object
        match = re.search(r"\{[\s\S]*\}", text)
        if not match:
            return {}

        try:
            data = json.loads(match.group(0))
            return data if isinstance(data, dict) else {}
        except json.JSONDecodeError:
            return {}

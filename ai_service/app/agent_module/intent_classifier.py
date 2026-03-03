"""Deterministic intent classifier for high-frequency shopping actions."""

from __future__ import annotations

import re
from dataclasses import dataclass, field
from typing import Any


SHOPPING_CATEGORIES = ("数码", "书籍", "生活", "服饰", "运动", "美妆", "乐器", "手游", "兼职", "其他")
ADD_TO_CART_HINT_RE = re.compile(
    r"加购|(加入|加进|加到|放入|放到|放进|塞进|添加到).{0,8}(购物车|购物袋|车里)|(购物车|购物袋|车里).{0,8}(加入|加进|加到|放入|放到|放进|塞进|添加到)"
)
ADD_TO_CART_DIRECT_ID_PATTERNS = (
    re.compile(
        r"^(?:帮我|请|麻烦|劳驾)?(?:把|将)?(?:商品)?(?P<id>\d{1,8})(?:号|号商品|商品|商品id|id)?"
        r"(?:加入|加进|加到|放入|放到|放进|塞进|添加到|加购)(?:到)?(?:我的|当前)?(?:购物车|购物袋|车里)?"
    ),
    re.compile(
        r"(?:加入|加进|加到|放入|放到|放进|塞进|添加到|加购)(?:到)?(?:我的|当前)?(?:购物车|购物袋|车里)?(?:里)?"
        r"(?:把|将)?(?P<id>\d{1,8})(?:号|号商品|商品|id)?"
    ),
)
ADD_TO_CART_NAME_PATTERNS = (
    re.compile(
        r"(?:帮我|请|麻烦|劳驾)?(?:把|将)?(?P<name>.+?)"
        r"(?:加入|加进|加到|放入|放到|放进|塞进|添加到|加购)(?:到)?(?:我的|当前)?(?:购物车|购物袋|车里)?$"
    ),
    re.compile(
        r"(?:帮我|请|麻烦|劳驾)?(?:加入|加进|加到|放入|放到|放进|塞进|添加到|加购)(?:到)?(?:我的|当前)?(?:购物车|购物袋|车里)?(?:里)?(?P<name>.+)$"
    ),
)
SEARCH_PREFIX_RE = re.compile(
    r"^(?:帮我|给我|替我|我想|想|请|麻烦|劳驾)?(?:搜一下|搜搜|搜|查一下|查下|查查|查|找一下|找找|找|看下|看看|看看有没有|看有没有|推荐一下|推荐点|推荐|来点|整点)(?P<term>.+)$"
)
NICKNAME_UPDATE_RE = re.compile(
    r"(?:昵称|名字|称呼)(?:帮我)?(?:改成|改为|换成|换为|设为|设置为|叫做|叫)(?P<value>[^，。！？!?]{1,24})"
)
NICKNAME_DIRECT_RE = re.compile(
    r"^(?:帮我|请|麻烦)?(?:把|将)?(?P<value>[^，。！？!?]{1,24})(?:设成|设为|改成|改为)(?:我的)?(?:昵称|名字|称呼)$"
)
PHONE_HINT_RE = re.compile(r"(手机号|电话|联系电话)")
AVATAR_HINT_RE = re.compile(r"(头像|照片|头像图)")
SETTING_HINT_RE = re.compile(r"(设置|资料|个人信息|个人资料|账号信息|账户信息|主页信息)")
ORDER_HINT_RE = re.compile(r"(订单|单子|物流|发货|收货|售后|退款|退单|购买记录)")
ORDER_QUERY_RE = re.compile(r"(查|看|看看|查询|状态|进度|物流|订单|到哪|在哪|咋样|怎么样)")
ORDER_BUYER_RE = re.compile(r"(我买的|买到的|买家订单|待付款|待发货|我的订单|购买记录|收货订单)")
ORDER_SELLER_RE = re.compile(r"(我卖的|卖出的|卖家订单|发出去的|卖货订单|卖掉的|作为卖家)")
CART_QUERY_RE = re.compile(r"(查看|看看|看下|查查|查下|查一下|列出|打开|瞅瞅|想看|购物车|购物袋|车里)")
FAVORITES_QUERY_RE = re.compile(r"(收藏|收藏夹|收藏的|心愿单|想要)")
GREETING_RE = re.compile(r"^(你好|您好|在吗|哈喽|hello|hi)\s*$", re.IGNORECASE)

LEADING_NOISE_RE = re.compile(
    r"^(?:把|将|这个|那个|这件|那件|这一件|那一件|这台|那台|这部|那部|这本|那本|这只|那只|一件|一个|一台|一部|一本|一只|一条|一双|我的|想把|我要把|我想把|我要)+"
)
TRAILING_NOISE_RE = re.compile(
    r"(?:加入购物车|加购物车|购物车|购物袋|车里|里|一下|吧|呀|呢|先|谢谢|哈|哦)+$"
)
STOPWORD_PRODUCT_NAMES = {"到", "里", "购物车", "购物袋", "车里", "它", "这个", "那个", "一下"}
QUERY_SUFFIX_RE = re.compile(r"(?:商品|宝贝|闲置|二手|好物|款)$")


@dataclass(frozen=True)
class ClarificationRequest:
    kind: str
    prompt: str
    options: tuple[str, ...] = ()
    context: dict[str, Any] = field(default_factory=dict)


@dataclass(frozen=True)
class IntentDecision:
    intent: str
    action: str
    confidence: float
    params: dict[str, Any] = field(default_factory=dict)
    clarification: ClarificationRequest | None = None
    message: str | None = None


class IntentClassifier:
    def classify(self, message: str, pending: dict[str, Any] | None = None) -> IntentDecision:
        text = self._normalize(message)
        if not text:
            return IntentDecision(intent="空输入", action="chat_only", confidence=0.3)

        if pending:
            resolved = self._resolve_pending(text, pending)
            if resolved is not None:
                return resolved

        if GREETING_RE.match(text):
            return IntentDecision(intent="打招呼", action="chat_only", confidence=0.98)

        add_to_cart = self._classify_add_to_cart(text)
        if add_to_cart is not None:
            return add_to_cart

        cart_query = self._classify_cart_query(text)
        if cart_query is not None:
            return cart_query

        favorites_query = self._classify_favorites_query(text)
        if favorites_query is not None:
            return favorites_query

        order_query = self._classify_order_query(text)
        if order_query is not None:
            return order_query

        profile_update = self._classify_profile_update(text)
        if profile_update is not None:
            return profile_update

        product_search = self._classify_product_search(text)
        if product_search is not None:
            return product_search

        return IntentDecision(intent="自由对话", action="chat_only", confidence=0.36)

    def _resolve_pending(self, text: str, pending: dict[str, Any]) -> IntentDecision | None:
        kind = str(pending.get("kind") or "").strip()
        if kind == "orders_scope":
            if ORDER_SELLER_RE.search(text):
                return IntentDecision(intent="查看卖家订单", action="orders.list", confidence=0.97, params={"role": "seller"})
            if ORDER_BUYER_RE.search(text) or "买的" in text or "我的" in text:
                return IntentDecision(intent="查看买家订单", action="orders.list", confidence=0.97, params={"role": "buyer"})
            return IntentDecision(
                intent="订单范围待确认",
                action="chat_only",
                confidence=0.42,
                clarification=ClarificationRequest(
                    kind="orders_scope",
                    prompt="您是想看我买到的订单，还是我卖出的订单？",
                    options=("买到的", "卖出的"),
                ),
            )

        if kind == "profile_field":
            nickname_decision = self._parse_nickname_update(text)
            if nickname_decision is not None:
                return nickname_decision
            if "昵称" in text or "名字" in text or "称呼" in text:
                return IntentDecision(
                    intent="昵称值待确认",
                    action="chat_only",
                    confidence=0.48,
                    clarification=ClarificationRequest(
                        kind="nickname_value",
                        prompt="您想改成什么昵称？直接回我新名字就行。",
                    ),
                )
            if AVATAR_HINT_RE.search(text):
                return IntentDecision(
                    intent="修改头像",
                    action="chat_only",
                    confidence=0.92,
                    message="头像需要在个人中心上传图片后保存。我这边能继续帮您改昵称，直接说“昵称改成小满”就行。",
                )
            if PHONE_HINT_RE.search(text):
                return IntentDecision(
                    intent="修改手机号",
                    action="chat_only",
                    confidence=0.92,
                    message="手机号建议在个人中心里修改。我能继续帮您改昵称，直接告诉我新的昵称即可。",
                )
            return IntentDecision(
                intent="资料字段待确认",
                action="chat_only",
                confidence=0.42,
                clarification=ClarificationRequest(
                    kind="profile_field",
                    prompt="您想改昵称、头像，还是手机号？",
                    options=("昵称", "头像", "手机号"),
                ),
            )

        if kind == "nickname_value":
            nickname_decision = self._parse_nickname_update(text, allow_plain_value=True)
            if nickname_decision is not None:
                return nickname_decision
            return IntentDecision(
                intent="昵称值待确认",
                action="chat_only",
                confidence=0.42,
                clarification=ClarificationRequest(
                    kind="nickname_value",
                    prompt="我还没拿到新昵称。直接回我一个名字，比如“小满旧物铺”。",
                ),
            )

        if kind == "add_to_cart_subject":
            add_to_cart = self._classify_add_to_cart(text)
            if add_to_cart is not None:
                return add_to_cart
            return IntentDecision(
                intent="商品名待确认",
                action="chat_only",
                confidence=0.4,
                clarification=ClarificationRequest(
                    kind="add_to_cart_subject",
                    prompt="您想把哪件商品加入购物车？直接说商品名，或者发商品 ID 也行。",
                ),
            )

        return None

    def _classify_add_to_cart(self, text: str) -> IntentDecision | None:
        compact = re.sub(r"\s+", "", text)
        if not ADD_TO_CART_HINT_RE.search(compact):
            return None

        for pattern in ADD_TO_CART_DIRECT_ID_PATTERNS:
            match = pattern.search(compact)
            if match:
                return IntentDecision(
                    intent="加入购物车",
                    action="cart.add",
                    confidence=0.99,
                    params={"product_id": int(match.group("id")), "count": 1},
                )

        product_name = ""
        for pattern in ADD_TO_CART_NAME_PATTERNS:
            match = pattern.search(text)
            if match:
                product_name = match.group("name")
                break

        if not product_name:
            quoted_match = re.search(r"[“\"'「『](.+?)[”\"'」』]", text)
            if quoted_match:
                product_name = quoted_match.group(1)

        product_name = self._clean_product_name(product_name)
        if product_name:
            if product_name.isdigit():
                return IntentDecision(
                    intent="加入购物车",
                    action="cart.add",
                    confidence=0.99,
                    params={"product_id": int(product_name), "count": 1},
                )
            return IntentDecision(
                intent="加入购物车",
                action="cart.add",
                confidence=0.94,
                params={"product_name": product_name, "count": 1},
            )

        return IntentDecision(
            intent="加入购物车",
            action="chat_only",
            confidence=0.48,
            clarification=ClarificationRequest(
                kind="add_to_cart_subject",
                prompt="您想把哪件商品加入购物车？直接说商品名，或者发商品 ID 也行。",
            ),
        )

    def _classify_cart_query(self, text: str) -> IntentDecision | None:
        if not re.search(r"(购物车|购物袋|车里)", text):
            return None
        if ADD_TO_CART_HINT_RE.search(text):
            return None
        if CART_QUERY_RE.search(text):
            return IntentDecision(intent="查看购物车", action="cart.list", confidence=0.95)
        return None

    def _classify_favorites_query(self, text: str) -> IntentDecision | None:
        if not FAVORITES_QUERY_RE.search(text):
            return None
        if re.search(r"(查看|看看|看下|查查|查一下|列出|翻翻|都有啥|都有什么|想看|了啥|有啥)", text):
            return IntentDecision(
                intent="查看收藏",
                action="user.data",
                confidence=0.93,
                params={"type": "favorites"},
            )
        return None

    def _classify_order_query(self, text: str) -> IntentDecision | None:
        if not ORDER_HINT_RE.search(text):
            return None
        if not ORDER_QUERY_RE.search(text):
            return None
        if ORDER_SELLER_RE.search(text):
            return IntentDecision(intent="查看卖家订单", action="orders.list", confidence=0.96, params={"role": "seller"})
        if ORDER_BUYER_RE.search(text):
            return IntentDecision(intent="查看买家订单", action="orders.list", confidence=0.93, params={"role": "buyer"})
        return IntentDecision(
            intent="订单查询",
            action="chat_only",
            confidence=0.46,
            clarification=ClarificationRequest(
                kind="orders_scope",
                prompt="您是想看我买到的订单，还是我卖出的订单？",
                options=("买到的", "卖出的"),
            ),
        )

    def _classify_profile_update(self, text: str) -> IntentDecision | None:
        nickname_decision = self._parse_nickname_update(text)
        if nickname_decision is not None:
            return nickname_decision

        if SETTING_HINT_RE.search(text) and re.search(r"(改|修改|设置|换|调整|动)", text):
            return IntentDecision(
                intent="修改资料",
                action="chat_only",
                confidence=0.44,
                clarification=ClarificationRequest(
                    kind="profile_field",
                    prompt="您想改昵称、头像，还是手机号？",
                    options=("昵称", "头像", "手机号"),
                ),
            )

        return None

    def _classify_product_search(self, text: str) -> IntentDecision | None:
        category_match = self._find_category(text)
        if category_match and re.search(r"(商品|宝贝|闲置|推荐|来点|找|搜|看|查|有没有)", text):
            return IntentDecision(
                intent="按分类搜商品",
                action="products.list",
                confidence=0.92,
                params={"category": category_match},
            )

        term = ""
        prefix_match = SEARCH_PREFIX_RE.search(text)
        if prefix_match:
            term = prefix_match.group("term")
        elif re.search(r"(有什么|有没有|来点)", text) and category_match:
            term = category_match

        term = self._clean_query_term(term)
        if not term:
            return None

        matched_category = self._find_category(term)
        if matched_category and term == matched_category:
            return IntentDecision(
                intent="按分类搜商品",
                action="products.list",
                confidence=0.92,
                params={"category": matched_category},
            )

        return IntentDecision(
          intent="搜索商品",
          action="products.list",
          confidence=0.9,
          params={"search": term},
        )

    def _parse_nickname_update(self, text: str, allow_plain_value: bool = False) -> IntentDecision | None:
        for pattern in (NICKNAME_UPDATE_RE, NICKNAME_DIRECT_RE):
            match = pattern.search(text)
            if match:
                value = self._clean_profile_value(match.group("value"))
                if value:
                    return IntentDecision(
                        intent="修改昵称",
                        action="user.profile.update",
                        confidence=0.96,
                        params={"nickname": value},
                    )

        if allow_plain_value:
            value = self._clean_profile_value(text)
            if value:
                return IntentDecision(
                    intent="修改昵称",
                    action="user.profile.update",
                    confidence=0.94,
                    params={"nickname": value},
                )
        return None

    def _find_category(self, text: str) -> str:
        for category in SHOPPING_CATEGORIES:
            if category in text:
                return category
        return ""

    def _clean_product_name(self, value: str) -> str:
        text = self._normalize(value)
        text = LEADING_NOISE_RE.sub("", text)
        text = TRAILING_NOISE_RE.sub("", text)
        text = text.strip("：:，,。.!！？?、 “”\"'‘’")
        if text in STOPWORD_PRODUCT_NAMES:
            return ""
        return text

    def _clean_query_term(self, value: str) -> str:
        text = self._normalize(value)
        text = re.sub(r"^(?:有没有|看有没有|看看有没有|给我|来个|来点)", "", text)
        text = re.sub(r"(给我|来个|来点|一下子|推荐点)$", "", text)
        text = text.strip("：:，,。.!！？?、 ")
        text = QUERY_SUFFIX_RE.sub("", text)
        return text.strip()

    def _clean_profile_value(self, value: str) -> str:
        text = self._normalize(value)
        text = text.strip("：:，,。.!！？?、 “”\"'‘’")
        if len(text) > 24:
            return ""
        if not text or re.search(r"(昵称|名字|称呼|设置|修改)$", text):
            return ""
        return text

    def _normalize(self, value: str) -> str:
        return re.sub(r"\s+", " ", str(value or "").strip())

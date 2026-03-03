from __future__ import annotations

import json
import sys
from collections import Counter, defaultdict
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from app.agent_module.intent_classifier import IntentClassifier  # noqa: E402


OUTPUT_DIR = Path(__file__).resolve().parent
CORPUS_PATH = OUTPUT_DIR / "intent_corpus.jsonl"
EVAL_CASES_PATH = OUTPUT_DIR / "intent_eval_cases.jsonl"
REPORT_PATH = OUTPUT_DIR / "intent_eval_report.md"


PRODUCT_NAMES = [
    "iPhone 15 Pro Max",
    "AirPods Pro 2",
    "Switch OLED",
    "索尼 WH-1000XM5",
    "小米 14 Ultra",
    "MacBook Air M3",
    "佳能 R50",
    "机械键盘 K8",
    "任天堂限定手柄",
    "戴森吹风机",
]

CATEGORIES = ["数码", "书籍", "生活", "服饰", "运动", "美妆", "乐器", "手游", "兼职", "其他"]
SEARCH_TERMS = [
    "露营灯",
    "二手相机",
    "羽毛球拍",
    "考研书",
    "复古耳机",
    "宿舍小冰箱",
    "黑胶唱片",
    "桌面音箱",
    "电竞鼠标",
    "法语教材",
]
NICKNAMES = [
    "旧物猎人",
    "阿梨杂货铺",
    "晚风寄卖站",
    "胶片不眠夜",
    "耳机收藏夹",
    "慢慢出闲置",
    "小满旧物铺",
    "周末换新",
    "纸箱藏宝图",
    "楼下拾光",
]


def write_jsonl(path: Path, rows: list[dict[str, Any]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as handle:
        for row in rows:
            handle.write(json.dumps(row, ensure_ascii=False) + "\n")


def generate_training_corpus() -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []

    add_cart_templates = [
        "把{name}加入购物车",
        "麻烦把{name}加到购物车",
        "请把{name}放进购物袋",
        "我要把{name}加购",
        "帮我把{name}塞进购物车",
        "把“{name}”加入到购物车",
        "将{name}添加到我的购物车",
        "麻烦把 {name} 加入车里",
        "请把{name}放到车里",
        "我想把{name}加入购物车",
    ]
    for name in PRODUCT_NAMES:
        for template in add_cart_templates:
            rows.append(
                {
                    "scenario": "add_to_cart_name",
                    "text": template.format(name=name),
                    "expected_action": "cart.add",
                    "expected_params": {"product_name": name, "count": 1},
                }
            )

    add_cart_id_templates = [
        "把{id}加入购物车",
        "帮我{id}加入到购物车",
        "请把商品{id}加入购物车",
        "将{id}号商品放进购物袋",
        "把{id}加购",
        "商品{id}放到车里",
    ]
    for product_id in range(101, 161):
        template = add_cart_id_templates[(product_id - 101) % len(add_cart_id_templates)]
        rows.append(
            {
                "scenario": "add_to_cart_id",
                "text": template.format(id=product_id),
                "expected_action": "cart.add",
                "expected_params": {"product_id": product_id, "count": 1},
            }
        )

    buyer_templates = [
        "帮我查一下我买的订单",
        "看看我的订单",
        "查下待付款订单",
        "我买到的订单现在啥状态",
        "帮我看看我的待发货订单",
        "查我买到的物流",
    ]
    for index in range(60):
        rows.append(
            {
                "scenario": "orders_buyer",
                "text": buyer_templates[index % len(buyer_templates)],
                "expected_action": "orders.list",
                "expected_params": {"role": "buyer"},
            }
        )

    seller_templates = [
        "帮我查一下我卖出的订单",
        "看看卖家订单",
        "查下我卖掉的订单",
        "我卖出的订单现在啥状态",
        "帮我看看待发货的卖家订单",
        "查我发出去的单子",
    ]
    for index in range(60):
        rows.append(
            {
                "scenario": "orders_seller",
                "text": seller_templates[index % len(seller_templates)],
                "expected_action": "orders.list",
                "expected_params": {"role": "seller"},
            }
        )

    ambiguous_order_templates = [
        "查订单",
        "看看订单",
        "订单状态怎么样",
        "帮我看下订单",
        "查一下物流",
    ]
    for index in range(40):
        rows.append(
            {
                "scenario": "orders_ambiguous",
                "text": ambiguous_order_templates[index % len(ambiguous_order_templates)],
                "expected_action": "chat_only",
                "expected_clarification_kind": "orders_scope",
            }
        )

    for category in CATEGORIES:
        for template in ("来点{category}商品", "看看{category}类宝贝", "有没有{category}闲置", "搜一下{category}"):
            rows.append(
                {
                    "scenario": "search_category",
                    "text": template.format(category=category),
                    "expected_action": "products.list",
                    "expected_params": {"category": category},
                }
            )

    for index in range(80):
        term = SEARCH_TERMS[index % len(SEARCH_TERMS)]
        template = ["帮我搜一下{term}", "找找{term}", "查下{term}", "推荐点{term}", "看看有没有{term}"][index % 5]
        rows.append(
            {
                "scenario": "search_keyword",
                "text": template.format(term=term),
                "expected_action": "products.list",
                "expected_params": {"search": term},
            }
        )

    cart_queries = [
        "看看我的购物车",
        "查一下购物车",
        "打开购物车",
        "购物车里有啥",
        "帮我列出购物车",
    ]
    for index in range(30):
        rows.append(
            {
                "scenario": "cart_query",
                "text": cart_queries[index % len(cart_queries)],
                "expected_action": "cart.list",
            }
        )

    favorites_queries = [
        "看看我的收藏",
        "查一下心愿单",
        "我都收藏了啥",
        "帮我列出想要的商品",
        "翻翻我的收藏夹",
    ]
    for index in range(30):
        rows.append(
            {
                "scenario": "favorites_query",
                "text": favorites_queries[index % len(favorites_queries)],
                "expected_action": "user.data",
                "expected_params": {"type": "favorites"},
            }
        )

    for nickname in NICKNAMES:
        for template in ("昵称改成{nickname}", "把我的昵称改为{nickname}", "名字换成{nickname}", "称呼设为{nickname}"):
            rows.append(
                {
                    "scenario": "nickname_update",
                    "text": template.format(nickname=nickname),
                    "expected_action": "user.profile.update",
                    "expected_params": {"nickname": nickname},
                }
            )

    settings_queries = [
        "帮我改一下资料",
        "我想修改个人信息",
        "设置一下账号资料",
        "帮我调整主页信息",
        "改改我的账户信息",
    ]
    for index in range(40):
        rows.append(
            {
                "scenario": "settings_ambiguous",
                "text": settings_queries[index % len(settings_queries)],
                "expected_action": "chat_only",
                "expected_clarification_kind": "profile_field",
            }
        )

    return rows


def generate_eval_cases() -> list[dict[str, Any]]:
    cases: list[dict[str, Any]] = []
    scenario_cases: dict[str, list[dict[str, Any]]] = {
        "add_to_cart_name": [
            {"text": "把iPhone 15 Pro Max加入购物车", "expected_action": "cart.add", "expected_params": {"product_name": "iPhone 15 Pro Max"}},
            {"text": "麻烦把 AirPods Pro 2 加到购物车", "expected_action": "cart.add", "expected_params": {"product_name": "AirPods Pro 2"}},
            {"text": "请把“Switch OLED”加入到购物车", "expected_action": "cart.add", "expected_params": {"product_name": "Switch OLED"}},
            {"text": "我要把索尼 WH-1000XM5加购", "expected_action": "cart.add", "expected_params": {"product_name": "索尼 WH-1000XM5"}},
            {"text": "把小米 14 Ultra 放进购物袋", "expected_action": "cart.add", "expected_params": {"product_name": "小米 14 Ultra"}},
            {"text": "将 MacBook Air M3 添加到我的购物车", "expected_action": "cart.add", "expected_params": {"product_name": "MacBook Air M3"}},
            {"text": "帮我把佳能 R50 塞进购物车", "expected_action": "cart.add", "expected_params": {"product_name": "佳能 R50"}},
            {"text": "请把机械键盘 K8 放到车里", "expected_action": "cart.add", "expected_params": {"product_name": "机械键盘 K8"}},
            {"text": "我想把任天堂限定手柄加入购物车", "expected_action": "cart.add", "expected_params": {"product_name": "任天堂限定手柄"}},
            {"text": "麻烦把戴森吹风机加到购物车", "expected_action": "cart.add", "expected_params": {"product_name": "戴森吹风机"}},
        ],
        "add_to_cart_id": [
            {"text": "把101加入购物车", "expected_action": "cart.add", "expected_params": {"product_id": 101}},
            {"text": "帮我102加入到购物车", "expected_action": "cart.add", "expected_params": {"product_id": 102}},
            {"text": "请把商品103加入购物车", "expected_action": "cart.add", "expected_params": {"product_id": 103}},
            {"text": "将104号商品放进购物袋", "expected_action": "cart.add", "expected_params": {"product_id": 104}},
            {"text": "把105加购", "expected_action": "cart.add", "expected_params": {"product_id": 105}},
            {"text": "商品106放到车里", "expected_action": "cart.add", "expected_params": {"product_id": 106}},
            {"text": "请把 107 加入购物车", "expected_action": "cart.add", "expected_params": {"product_id": 107}},
            {"text": "帮我把108加到购物袋", "expected_action": "cart.add", "expected_params": {"product_id": 108}},
            {"text": "把109号商品加进购物车", "expected_action": "cart.add", "expected_params": {"product_id": 109}},
            {"text": "将110加入到我的购物车", "expected_action": "cart.add", "expected_params": {"product_id": 110}},
        ],
        "orders_buyer": [
            {"text": "帮我查一下我买的订单", "expected_action": "orders.list", "expected_params": {"role": "buyer"}},
            {"text": "看看我的订单", "expected_action": "orders.list", "expected_params": {"role": "buyer"}},
            {"text": "查下待付款订单", "expected_action": "orders.list", "expected_params": {"role": "buyer"}},
            {"text": "我买到的订单现在啥状态", "expected_action": "orders.list", "expected_params": {"role": "buyer"}},
            {"text": "帮我看看我的待发货订单", "expected_action": "orders.list", "expected_params": {"role": "buyer"}},
            {"text": "查我买到的物流", "expected_action": "orders.list", "expected_params": {"role": "buyer"}},
            {"text": "我买的单子到哪了", "expected_action": "orders.list", "expected_params": {"role": "buyer"}},
            {"text": "帮我翻一下买家订单", "expected_action": "orders.list", "expected_params": {"role": "buyer"}},
            {"text": "看下我的收货订单", "expected_action": "orders.list", "expected_params": {"role": "buyer"}},
            {"text": "查查我的购买记录", "expected_action": "orders.list", "expected_params": {"role": "buyer"}},
        ],
        "orders_seller": [
            {"text": "帮我查一下我卖出的订单", "expected_action": "orders.list", "expected_params": {"role": "seller"}},
            {"text": "看看卖家订单", "expected_action": "orders.list", "expected_params": {"role": "seller"}},
            {"text": "查下我卖掉的订单", "expected_action": "orders.list", "expected_params": {"role": "seller"}},
            {"text": "我卖出的订单现在啥状态", "expected_action": "orders.list", "expected_params": {"role": "seller"}},
            {"text": "帮我看看待发货的卖家订单", "expected_action": "orders.list", "expected_params": {"role": "seller"}},
            {"text": "查我发出去的单子", "expected_action": "orders.list", "expected_params": {"role": "seller"}},
            {"text": "卖出的订单都怎么样了", "expected_action": "orders.list", "expected_params": {"role": "seller"}},
            {"text": "帮我翻一下卖货订单", "expected_action": "orders.list", "expected_params": {"role": "seller"}},
            {"text": "看下我作为卖家的订单", "expected_action": "orders.list", "expected_params": {"role": "seller"}},
            {"text": "查查我卖出的物流", "expected_action": "orders.list", "expected_params": {"role": "seller"}},
        ],
        "orders_ambiguous": [
            {"text": "查订单", "expected_action": "chat_only", "expected_clarification_kind": "orders_scope"},
            {"text": "看看订单", "expected_action": "chat_only", "expected_clarification_kind": "orders_scope"},
            {"text": "订单状态怎么样", "expected_action": "chat_only", "expected_clarification_kind": "orders_scope"},
            {"text": "帮我看下订单", "expected_action": "chat_only", "expected_clarification_kind": "orders_scope"},
            {"text": "查一下物流", "expected_action": "chat_only", "expected_clarification_kind": "orders_scope"},
            {"text": "订单现在到哪步了", "expected_action": "chat_only", "expected_clarification_kind": "orders_scope"},
            {"text": "帮我看看最近的单子", "expected_action": "chat_only", "expected_clarification_kind": "orders_scope"},
            {"text": "我想查订单进度", "expected_action": "chat_only", "expected_clarification_kind": "orders_scope"},
            {"text": "查查最近物流", "expected_action": "chat_only", "expected_clarification_kind": "orders_scope"},
            {"text": "看看单子", "expected_action": "chat_only", "expected_clarification_kind": "orders_scope"},
        ],
        "search_category": [
            {"text": "来点数码商品", "expected_action": "products.list", "expected_params": {"category": "数码"}},
            {"text": "看看书籍类宝贝", "expected_action": "products.list", "expected_params": {"category": "书籍"}},
            {"text": "有没有生活闲置", "expected_action": "products.list", "expected_params": {"category": "生活"}},
            {"text": "搜一下服饰", "expected_action": "products.list", "expected_params": {"category": "服饰"}},
            {"text": "来点运动商品", "expected_action": "products.list", "expected_params": {"category": "运动"}},
            {"text": "看看美妆类宝贝", "expected_action": "products.list", "expected_params": {"category": "美妆"}},
            {"text": "有没有乐器闲置", "expected_action": "products.list", "expected_params": {"category": "乐器"}},
            {"text": "搜一下手游", "expected_action": "products.list", "expected_params": {"category": "手游"}},
            {"text": "来点兼职信息", "expected_action": "products.list", "expected_params": {"category": "兼职"}},
            {"text": "看看其他类商品", "expected_action": "products.list", "expected_params": {"category": "其他"}},
        ],
        "search_keyword": [
            {"text": "帮我搜一下露营灯", "expected_action": "products.list", "expected_params": {"search": "露营灯"}},
            {"text": "找找二手相机", "expected_action": "products.list", "expected_params": {"search": "二手相机"}},
            {"text": "查下羽毛球拍", "expected_action": "products.list", "expected_params": {"search": "羽毛球拍"}},
            {"text": "推荐点考研书", "expected_action": "products.list", "expected_params": {"search": "考研书"}},
            {"text": "看看有没有复古耳机", "expected_action": "products.list", "expected_params": {"search": "复古耳机"}},
            {"text": "帮我搜一下宿舍小冰箱", "expected_action": "products.list", "expected_params": {"search": "宿舍小冰箱"}},
            {"text": "找找黑胶唱片", "expected_action": "products.list", "expected_params": {"search": "黑胶唱片"}},
            {"text": "查下桌面音箱", "expected_action": "products.list", "expected_params": {"search": "桌面音箱"}},
            {"text": "推荐点电竞鼠标", "expected_action": "products.list", "expected_params": {"search": "电竞鼠标"}},
            {"text": "看看有没有法语教材", "expected_action": "products.list", "expected_params": {"search": "法语教材"}},
        ],
        "cart_query": [
            {"text": "看看我的购物车", "expected_action": "cart.list"},
            {"text": "查一下购物车", "expected_action": "cart.list"},
            {"text": "打开购物车", "expected_action": "cart.list"},
            {"text": "购物车里有啥", "expected_action": "cart.list"},
            {"text": "帮我列出购物车", "expected_action": "cart.list"},
            {"text": "查查我的购物袋", "expected_action": "cart.list"},
            {"text": "让我看看车里有什么", "expected_action": "cart.list"},
            {"text": "把购物车打开", "expected_action": "cart.list"},
            {"text": "我购物车现在啥情况", "expected_action": "cart.list"},
            {"text": "想看购物车内容", "expected_action": "cart.list"},
        ],
        "favorites_query": [
            {"text": "看看我的收藏", "expected_action": "user.data", "expected_params": {"type": "favorites"}},
            {"text": "查一下心愿单", "expected_action": "user.data", "expected_params": {"type": "favorites"}},
            {"text": "我都收藏了啥", "expected_action": "user.data", "expected_params": {"type": "favorites"}},
            {"text": "帮我列出想要的商品", "expected_action": "user.data", "expected_params": {"type": "favorites"}},
            {"text": "翻翻我的收藏夹", "expected_action": "user.data", "expected_params": {"type": "favorites"}},
            {"text": "看下我收藏的东西", "expected_action": "user.data", "expected_params": {"type": "favorites"}},
            {"text": "我的心愿单里有啥", "expected_action": "user.data", "expected_params": {"type": "favorites"}},
            {"text": "帮我查查收藏", "expected_action": "user.data", "expected_params": {"type": "favorites"}},
            {"text": "想看看收藏列表", "expected_action": "user.data", "expected_params": {"type": "favorites"}},
            {"text": "查一下我想要的宝贝", "expected_action": "user.data", "expected_params": {"type": "favorites"}},
        ],
        "nickname_update": [
            {"text": "昵称改成旧物猎人", "expected_action": "user.profile.update", "expected_params": {"nickname": "旧物猎人"}},
            {"text": "把我的昵称改为阿梨杂货铺", "expected_action": "user.profile.update", "expected_params": {"nickname": "阿梨杂货铺"}},
            {"text": "名字换成晚风寄卖站", "expected_action": "user.profile.update", "expected_params": {"nickname": "晚风寄卖站"}},
            {"text": "称呼设为胶片不眠夜", "expected_action": "user.profile.update", "expected_params": {"nickname": "胶片不眠夜"}},
            {"text": "昵称改成耳机收藏夹", "expected_action": "user.profile.update", "expected_params": {"nickname": "耳机收藏夹"}},
            {"text": "把我的昵称改为慢慢出闲置", "expected_action": "user.profile.update", "expected_params": {"nickname": "慢慢出闲置"}},
            {"text": "名字换成小满旧物铺", "expected_action": "user.profile.update", "expected_params": {"nickname": "小满旧物铺"}},
            {"text": "称呼设为周末换新", "expected_action": "user.profile.update", "expected_params": {"nickname": "周末换新"}},
            {"text": "昵称改成纸箱藏宝图", "expected_action": "user.profile.update", "expected_params": {"nickname": "纸箱藏宝图"}},
            {"text": "把我的昵称改为楼下拾光", "expected_action": "user.profile.update", "expected_params": {"nickname": "楼下拾光"}},
        ],
        "settings_ambiguous": [
            {"text": "帮我改一下资料", "expected_action": "chat_only", "expected_clarification_kind": "profile_field"},
            {"text": "我想修改个人信息", "expected_action": "chat_only", "expected_clarification_kind": "profile_field"},
            {"text": "设置一下账号资料", "expected_action": "chat_only", "expected_clarification_kind": "profile_field"},
            {"text": "帮我调整主页信息", "expected_action": "chat_only", "expected_clarification_kind": "profile_field"},
            {"text": "改改我的账户信息", "expected_action": "chat_only", "expected_clarification_kind": "profile_field"},
            {"text": "我想改一下个人资料", "expected_action": "chat_only", "expected_clarification_kind": "profile_field"},
            {"text": "帮我重新设置资料", "expected_action": "chat_only", "expected_clarification_kind": "profile_field"},
            {"text": "账号信息想调整一下", "expected_action": "chat_only", "expected_clarification_kind": "profile_field"},
            {"text": "个人主页信息得改改", "expected_action": "chat_only", "expected_clarification_kind": "profile_field"},
            {"text": "帮我动一下资料", "expected_action": "chat_only", "expected_clarification_kind": "profile_field"},
        ],
    }

    for scenario, rows in scenario_cases.items():
        for index, row in enumerate(rows, start=1):
            cases.append({"id": f"{scenario}-{index:02d}", "scenario": scenario, **row})
    return cases


def params_match(expected: dict[str, Any], actual: dict[str, Any]) -> bool:
    for key, value in expected.items():
        if actual.get(key) != value:
            return False
    return True


def evaluate_cases(cases: list[dict[str, Any]]) -> tuple[list[dict[str, Any]], dict[str, Any]]:
    classifier = IntentClassifier()
    results: list[dict[str, Any]] = []
    scenario_stats: dict[str, dict[str, int]] = defaultdict(lambda: {"total": 0, "passed": 0})

    for case in cases:
        decision = classifier.classify(case["text"])
        passed = decision.action == case["expected_action"]
        reason = ""

        expected_params = case.get("expected_params") or {}
        if passed and expected_params and not params_match(expected_params, decision.params):
            passed = False
            reason = "params_mismatch"

        expected_clarification_kind = case.get("expected_clarification_kind")
        actual_clarification_kind = decision.clarification.kind if decision.clarification else ""
        if passed and expected_clarification_kind and actual_clarification_kind != expected_clarification_kind:
            passed = False
            reason = "clarification_kind_mismatch"

        if passed and expected_clarification_kind and decision.confidence >= 0.8:
            passed = False
            reason = "clarification_confidence_too_high"

        if passed and not expected_clarification_kind and decision.confidence < 0.8:
            passed = False
            reason = "direct_confidence_too_low"

        scenario_stats[case["scenario"]]["total"] += 1
        if passed:
            scenario_stats[case["scenario"]]["passed"] += 1

        results.append(
            {
                "id": case["id"],
                "scenario": case["scenario"],
                "text": case["text"],
                "passed": passed,
                "expected_action": case["expected_action"],
                "actual_action": decision.action,
                "expected_params": expected_params,
                "actual_params": decision.params,
                "expected_clarification_kind": expected_clarification_kind or "",
                "actual_clarification_kind": actual_clarification_kind,
                "confidence": round(decision.confidence, 4),
                "reason": reason,
            }
        )

    total = len(results)
    passed_total = sum(1 for row in results if row["passed"])
    summary = {
        "total": total,
        "passed": passed_total,
        "accuracy": round((passed_total / total) * 100, 2) if total else 0.0,
        "scenario_stats": scenario_stats,
    }
    return results, summary


def build_report(corpus_rows: list[dict[str, Any]], eval_results: list[dict[str, Any]], summary: dict[str, Any]) -> str:
    failed_rows = [row for row in eval_results if not row["passed"]]
    scenario_lines = []
    for scenario, stats in summary["scenario_stats"].items():
        accuracy = (stats["passed"] / stats["total"] * 100) if stats["total"] else 0
        scenario_lines.append(f"| {scenario} | {stats['passed']} / {stats['total']} | {accuracy:.2f}% |")

    corpus_counter = Counter(row["scenario"] for row in corpus_rows)
    corpus_lines = []
    for scenario, count in sorted(corpus_counter.items()):
        corpus_lines.append(f"| {scenario} | {count} |")

    failure_lines = []
    if failed_rows:
        for row in failed_rows[:12]:
            failure_lines.append(
                f"| {row['id']} | {row['text']} | {row['expected_action']} | {row['actual_action']} | {row['reason']} |"
            )
    else:
        failure_lines.append("| 无 | - | - | - | - |")

    return "\n".join(
        [
            "# Intent Eval Report",
            "",
            "## Summary",
            "",
            f"- 训练就绪语料: {len(corpus_rows)} 条",
            f"- 回归测试集: {summary['total']} 条",
            f"- 通过数: {summary['passed']}",
            f"- 准确率: {summary['accuracy']:.2f}%",
            f"- 说明: 仓库里没有历史标注基线，本报告只统计优化后的 {summary['total']} 条回归集结果。",
            "",
            "## Scenario Breakdown",
            "",
            "| 场景 | 通过 / 总数 | 准确率 |",
            "| --- | --- | --- |",
            *scenario_lines,
            "",
            "## Corpus Breakdown",
            "",
            "| 场景 | 样本数 |",
            "| --- | --- |",
            *corpus_lines,
            "",
            "## Hard Cases Covered",
            "",
            "1. 商品名里有空格、英文和数字混排",
            "2. 纯数字商品 ID 加购",
            "3. 买家订单和卖家订单混淆",
            "4. 只说“查订单”时先追问，不直接乱执行",
            "5. 只说“改资料”时先追问字段",
            "6. 分类搜索和关键词搜索拆开判断",
            "7. 收藏查询和购物车查询不串台",
            "8. 昵称修改只在值明确时执行",
            "9. 口语化短句和省略主语",
            "10. 带引号的商品名",
            "",
            "## Failed Cases",
            "",
            "| 用例 | 文本 | 期望动作 | 实际动作 | 原因 |",
            "| --- | --- | --- | --- | --- |",
            *failure_lines,
            "",
        ]
    )


def main() -> None:
    corpus_rows = generate_training_corpus()
    eval_cases = generate_eval_cases()
    eval_results, summary = evaluate_cases(eval_cases)

    write_jsonl(CORPUS_PATH, corpus_rows)
    write_jsonl(EVAL_CASES_PATH, eval_cases)
    REPORT_PATH.write_text(build_report(corpus_rows, eval_results, summary), encoding="utf-8")

    print(json.dumps({"corpus": len(corpus_rows), **summary}, ensure_ascii=False))


if __name__ == "__main__":
    main()

# Intent Eval Report

## Summary

- 训练就绪语料: 580 条
- 回归测试集: 110 条
- 通过数: 110
- 准确率: 100.00%
- 说明: 仓库里没有历史标注基线，本报告只统计优化后的 110 条回归集结果。

## Scenario Breakdown

| 场景 | 通过 / 总数 | 准确率 |
| --- | --- | --- |
| add_to_cart_name | 10 / 10 | 100.00% |
| add_to_cart_id | 10 / 10 | 100.00% |
| orders_buyer | 10 / 10 | 100.00% |
| orders_seller | 10 / 10 | 100.00% |
| orders_ambiguous | 10 / 10 | 100.00% |
| search_category | 10 / 10 | 100.00% |
| search_keyword | 10 / 10 | 100.00% |
| cart_query | 10 / 10 | 100.00% |
| favorites_query | 10 / 10 | 100.00% |
| nickname_update | 10 / 10 | 100.00% |
| settings_ambiguous | 10 / 10 | 100.00% |

## Corpus Breakdown

| 场景 | 样本数 |
| --- | --- |
| add_to_cart_id | 60 |
| add_to_cart_name | 100 |
| cart_query | 30 |
| favorites_query | 30 |
| nickname_update | 40 |
| orders_ambiguous | 40 |
| orders_buyer | 60 |
| orders_seller | 60 |
| search_category | 40 |
| search_keyword | 80 |
| settings_ambiguous | 40 |

## Hard Cases Covered

1. 商品名里有空格、英文和数字混排
2. 纯数字商品 ID 加购
3. 买家订单和卖家订单混淆
4. 只说“查订单”时先追问，不直接乱执行
5. 只说“改资料”时先追问字段
6. 分类搜索和关键词搜索拆开判断
7. 收藏查询和购物车查询不串台
8. 昵称修改只在值明确时执行
9. 口语化短句和省略主语
10. 带引号的商品名

## Failed Cases

| 用例 | 文本 | 期望动作 | 实际动作 | 原因 |
| --- | --- | --- | --- | --- |
| 无 | - | - | - | - |

from __future__ import annotations

import json
import os
from collections import Counter
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
LOG_DIR = Path(os.getenv("AGENT_TELEMETRY_DIR", str(ROOT / "logs" / "agent")))
LOG_FILE = LOG_DIR / "turns.jsonl"
REPORT_PATH = Path(__file__).resolve().parent / "agent_log_report.md"


def load_rows() -> list[dict]:
    if not LOG_FILE.exists():
        return []
    rows: list[dict] = []
    for line in LOG_FILE.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            data = json.loads(line)
        except json.JSONDecodeError:
            continue
        if isinstance(data, dict):
            rows.append(data)
    return rows


def build_report(rows: list[dict]) -> str:
    if not rows:
        return "\n".join(
            [
                "# Agent Log Report",
                "",
                "当前还没有采集到 Agent 对话日志。",
                "",
                f"- 预期日志文件: `{LOG_FILE}`",
                "- 先让线上 Agent 跑一段时间，再重新执行这个脚本。",
                "",
            ]
        )

    intents = Counter(str(row.get("intent") or "未命中") for row in rows)
    sources = Counter(str(row.get("source") or "unknown") for row in rows)
    sentiments = Counter(str(row.get("sentiment") or "neutral") for row in rows)
    service_modes = Counter(str(row.get("service_mode") or "guide") for row in rows)
    fallback_count = sum(1 for row in rows if row.get("fallback"))

    top_intents = [f"| {name} | {count} |" for name, count in intents.most_common(10)]
    top_sources = [f"| {name} | {count} |" for name, count in sources.most_common()]
    top_sentiments = [f"| {name} | {count} |" for name, count in sentiments.most_common()]
    top_modes = [f"| {name} | {count} |" for name, count in service_modes.most_common()]

    return "\n".join(
        [
            "# Agent Log Report",
            "",
            f"- 总对话轮次: {len(rows)}",
            f"- fallback 轮次: {fallback_count}",
            "",
            "## 高频意图",
            "",
            "| 意图 | 次数 |",
            "| --- | --- |",
            *(top_intents or ["| 无 | 0 |"]),
            "",
            "## 来源分布",
            "",
            "| 来源 | 次数 |",
            "| --- | --- |",
            *(top_sources or ["| 无 | 0 |"]),
            "",
            "## 情绪分布",
            "",
            "| 情绪 | 次数 |",
            "| --- | --- |",
            *(top_sentiments or ["| 无 | 0 |"]),
            "",
            "## 服务模式分布",
            "",
            "| 模式 | 次数 |",
            "| --- | --- |",
            *(top_modes or ["| 无 | 0 |"]),
            "",
        ]
    )


def main() -> None:
    rows = load_rows()
    REPORT_PATH.write_text(build_report(rows), encoding="utf-8")
    print(f"rows={len(rows)} report={REPORT_PATH}")


if __name__ == "__main__":
    main()

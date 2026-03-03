"""Agent telemetry helpers."""

from __future__ import annotations

import json
import re
from datetime import datetime, timezone
from pathlib import Path
from threading import Lock
from typing import Any


EMAIL_RE = re.compile(r"([A-Za-z0-9._%+-]{2})[A-Za-z0-9._%+-]*@([A-Za-z0-9.-]+\.[A-Za-z]{2,})")
PHONE_RE = re.compile(r"(?<!\d)(1[3-9]\d{2})\d{4}(\d{4})(?!\d)")
TOKEN_RE = re.compile(r"\b(?:Bearer\s+)?[A-Za-z0-9_\-]{24,}\b")


def mask_sensitive_text(raw: str) -> str:
    text = str(raw or "").strip()
    if not text:
        return ""
    text = EMAIL_RE.sub(r"\1***@\2", text)
    text = PHONE_RE.sub(r"\1****\2", text)
    text = TOKEN_RE.sub("[redacted-token]", text)
    return text


class AgentTelemetryLogger:
    def __init__(self, directory: str) -> None:
        self._directory = Path(directory)
        self._directory.mkdir(parents=True, exist_ok=True)
        self._path = self._directory / "turns.jsonl"
        self._lock = Lock()

    def log_turn(self, payload: dict[str, Any]) -> None:
        record = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            **payload,
        }
        if "user_text" in record:
            record["user_text"] = mask_sensitive_text(str(record["user_text"]))
        if "answer" in record:
            record["answer"] = mask_sensitive_text(str(record["answer"]))

        line = json.dumps(record, ensure_ascii=False)
        with self._lock:
            with self._path.open("a", encoding="utf-8") as handle:
                handle.write(line + "\n")

from __future__ import annotations

import json
import os
from datetime import datetime, timezone
from pathlib import Path

import httpx


ROOT = Path(__file__).resolve().parents[1]
README_PATH = ROOT.parent / "README.md"
OUTPUT_PATH = ROOT / "knowledge" / "agent_knowledge_snapshot.json"
BACKEND_API_BASE_URL = os.getenv("BACKEND_API_BASE_URL", "https://api.315279.xyz/api").rstrip("/")


def clean_markdown_line(raw_line: str) -> str:
    line = str(raw_line or "").strip()
    if not line or line in {"---", "***"} or line.startswith("```"):
        return ""
    line = line.replace("`", "")
    line = line.lstrip("-* ")
    line = " ".join(line.split())
    return line


def read_readme_summary() -> list[dict[str, str]]:
    if not README_PATH.exists():
        return []

    faq_entries: list[dict[str, str]] = []
    current_title = ""
    bucket: list[str] = []
    for raw_line in README_PATH.read_text(encoding="utf-8").splitlines():
        line = clean_markdown_line(raw_line)
        if not line:
            continue
        if line.startswith("## "):
            if current_title and bucket:
                faq_entries.append({"title": current_title, "content": " ".join(bucket[:4])})
            current_title = line.removeprefix("## ").strip()
            bucket = []
            continue
        if current_title and not line.startswith("#") and not line.startswith("|"):
            bucket.append(line)

    if current_title and bucket:
        faq_entries.append({"title": current_title, "content": " ".join(bucket[:4])})
    return faq_entries[:12]


def fetch_public_payload(path: str, params: dict | None = None) -> dict | list:
    with httpx.Client(timeout=20, trust_env=False) as client:
        response = client.get(f"{BACKEND_API_BASE_URL}{path}", params=params)
        response.raise_for_status()
        return response.json()


def sync_snapshot() -> dict:
    categories = fetch_public_payload("/categories")
    products = fetch_public_payload("/products", {"page": 1, "page_size": 30})
    faq_entries = read_readme_summary()
    category_items: list[dict] = []

    if isinstance(categories, list):
        category_items = [item for item in categories if isinstance(item, dict)]
    elif isinstance(categories, dict):
        raw_categories = categories.get("data") or categories.get("list") or categories.get("items") or []
        if isinstance(raw_categories, list):
            category_items = [item for item in raw_categories if isinstance(item, dict)]

    items = []
    if isinstance(products, dict):
        raw_items = products.get("list") or products.get("data") or []
        if isinstance(raw_items, list):
            for item in raw_items[:20]:
                if not isinstance(item, dict):
                    continue
                items.append(
                    {
                        "id": item.get("id"),
                        "name": item.get("name"),
                        "category": item.get("category"),
                        "price": item.get("price"),
                        "description": str(item.get("description") or "")[:160],
                    }
                )

    snapshot = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "backend_api_base_url": BACKEND_API_BASE_URL,
        "faq_entries": faq_entries,
        "categories": category_items[:20],
        "products": items,
    }
    OUTPUT_PATH.parent.mkdir(parents=True, exist_ok=True)
    OUTPUT_PATH.write_text(json.dumps(snapshot, ensure_ascii=False, indent=2), encoding="utf-8")
    return snapshot


def main() -> None:
    snapshot = sync_snapshot()
    print(json.dumps({"faq_entries": len(snapshot["faq_entries"]), "products": len(snapshot["products"])}, ensure_ascii=False))


if __name__ == "__main__":
    main()

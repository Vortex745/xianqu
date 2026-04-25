from __future__ import annotations

from typing import Iterable


def normalize_top_k(top_k: int, default: int = 20) -> int:
    if top_k <= 0:
        return default
    return min(top_k, 50)


def dedupe_candidate_products(products: Iterable[dict]) -> list[dict]:
    result: list[dict] = []
    seen_ids: set[int] = set()
    for raw in products:
        pid = int(raw.get("id", 0) or 0)
        if pid <= 0 or pid in seen_ids:
            continue
        seen_ids.add(pid)
        result.append(
            {
                "id": pid,
                "name": str(raw.get("name", "") or "").strip(),
                "category": str(raw.get("category", "") or "").strip(),
                "price": float(raw.get("price", 0) or 0),
                "view_count": int(raw.get("view_count", 0) or 0),
                "popularity": float(raw.get("popularity", 0) or 0),
            }
        )
    return result


def popularity_scores(products: Iterable[dict]) -> dict[int, float]:
    scores: dict[int, float] = {}
    for item in products:
        pid = int(item.get("id", 0) or 0)
        if pid <= 0:
            continue
        popularity = float(item.get("popularity", 0) or 0)
        view_count = float(item.get("view_count", 0) or 0)
        scores[pid] = max(popularity, view_count)
    return scores


def normalize_score_map(score_map: dict[int, float]) -> dict[int, float]:
    if not score_map:
        return {}
    values = [float(v) for v in score_map.values()]
    min_v = min(values)
    max_v = max(values)
    if max_v - min_v < 1e-9:
        return {k: 0.0 for k in score_map}
    return {k: (float(v) - min_v) / (max_v - min_v) for k, v in score_map.items()}


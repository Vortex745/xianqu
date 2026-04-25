from __future__ import annotations

from typing import Any, Iterable

from .cf_engine import compute_cf_scores
from .content_engine import compute_content_scores
from .data_loader import (
    dedupe_candidate_products,
    normalize_score_map,
    normalize_top_k,
    popularity_scores,
)


class RecommenderService:
    def recommend(
        self,
        user_id: int,
        user_item_scores: Iterable[Any],
        candidate_products: Iterable[Any],
        top_k: int = 20,
        behavior_rows: Iterable[Any] | None = None,
    ) -> tuple[list[int], str]:
        top_k = normalize_top_k(int(top_k or 0))
        candidates = dedupe_candidate_products(self._to_dict_list(candidate_products))
        if not candidates:
            return [], "hot"

        normalized_user_scores = self._normalize_user_item_scores(user_item_scores)
        popularity_map = popularity_scores(candidates)
        candidate_ids = {int(item["id"]) for item in candidates}

        # 冷启动：无用户行为，直接走热门
        if user_id <= 0 or not normalized_user_scores:
            return self._hot_candidates(candidates, top_k), "hot"

        normalized_behavior_rows = self._normalize_behavior_rows(behavior_rows or [])

        cf_scores = compute_cf_scores(
            target_user_id=int(user_id),
            behavior_rows=normalized_behavior_rows,
            candidate_product_ids=candidate_ids,
        )
        content_scores = compute_content_scores(
            user_item_scores=normalized_user_scores,
            candidate_products=candidates,
        )

        normalized_cf = normalize_score_map(cf_scores)
        normalized_content = normalize_score_map(content_scores)
        normalized_popularity = normalize_score_map(popularity_map)

        combined: dict[int, float] = {}
        for pid in candidate_ids:
            combined[pid] = (
                0.65 * normalized_cf.get(pid, 0.0)
                + 0.35 * normalized_content.get(pid, 0.0)
                + 0.08 * normalized_popularity.get(pid, 0.0)
            )

        ranked = sorted(
            combined.items(),
            key=lambda item: (item[1], normalized_popularity.get(item[0], 0.0)),
            reverse=True,
        )
        product_ids = [pid for pid, score in ranked if score > 1e-9][:top_k]

        if not product_ids:
            return self._hot_candidates(candidates, top_k), "hot"
        return product_ids, "personalized"

    def _to_dict_list(self, values: Iterable[Any]) -> list[dict]:
        result: list[dict] = []
        for value in values:
            if isinstance(value, dict):
                result.append(value)
                continue
            if hasattr(value, "model_dump"):
                result.append(value.model_dump())
                continue
            if hasattr(value, "dict"):
                result.append(value.dict())
                continue
        return result

    def _normalize_user_item_scores(self, rows: Iterable[Any]) -> list[dict]:
        result: list[dict] = []
        for row in self._to_dict_list(rows):
            product_id = int(row.get("product_id", 0) or 0)
            score = float(row.get("score", 0) or 0)
            if product_id <= 0 or score <= 0:
                continue
            result.append({"product_id": product_id, "score": score})
        return result

    def _normalize_behavior_rows(self, rows: Iterable[Any]) -> list[dict]:
        result: list[dict] = []
        for row in self._to_dict_list(rows):
            user_id = int(row.get("user_id", 0) or 0)
            product_id = int(row.get("product_id", 0) or 0)
            weight = float(row.get("weight", 0) or 0)
            if user_id <= 0 or product_id <= 0 or weight <= 0:
                continue
            result.append({"user_id": user_id, "product_id": product_id, "weight": weight})
        return result

    def _hot_candidates(self, candidates: list[dict], top_k: int) -> list[int]:
        ranked = sorted(
            candidates,
            key=lambda item: (
                float(item.get("popularity", 0) or 0),
                float(item.get("view_count", 0) or 0),
            ),
            reverse=True,
        )
        return [int(item["id"]) for item in ranked[:top_k]]


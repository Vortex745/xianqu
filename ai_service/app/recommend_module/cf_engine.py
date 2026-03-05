from __future__ import annotations

from typing import Iterable

try:
    import numpy as np
    from scipy import sparse
    from sklearn.metrics.pairwise import cosine_similarity
except Exception:  # pragma: no cover - graceful degradation when deps missing
    np = None
    sparse = None
    cosine_similarity = None


def compute_cf_scores(
    target_user_id: int,
    behavior_rows: Iterable[dict],
    candidate_product_ids: set[int],
) -> dict[int, float]:
    if not target_user_id or not candidate_product_ids:
        return {}
    if np is None or sparse is None or cosine_similarity is None:
        return {}

    rows = [row for row in behavior_rows if int(row.get("product_id", 0) or 0) > 0 and int(row.get("user_id", 0) or 0) > 0]
    if not rows:
        return {}

    user_ids = sorted({int(row["user_id"]) for row in rows})
    item_ids = sorted({int(row["product_id"]) for row in rows})
    if target_user_id not in user_ids:
        return {}

    user_to_idx = {uid: idx for idx, uid in enumerate(user_ids)}
    item_to_idx = {pid: idx for idx, pid in enumerate(item_ids)}

    matrix_rows: list[int] = []
    matrix_cols: list[int] = []
    matrix_data: list[float] = []

    for row in rows:
        uid = int(row["user_id"])
        pid = int(row["product_id"])
        matrix_rows.append(user_to_idx[uid])
        matrix_cols.append(item_to_idx[pid])
        matrix_data.append(float(row.get("weight", 0) or 0))

    if not matrix_data:
        return {}

    interaction_matrix = sparse.csr_matrix(
        (matrix_data, (matrix_rows, matrix_cols)),
        shape=(len(user_ids), len(item_ids)),
        dtype=float,
    )

    target_idx = user_to_idx[target_user_id]
    target_vector = interaction_matrix[target_idx]
    if target_vector.nnz == 0:
        return {}

    # 用户相似度 -> 预测目标用户对各商品的偏好
    user_similarity = cosine_similarity(target_vector, interaction_matrix).flatten()
    if user_similarity.sum() == 0:
        return {}

    predicted_scores = (user_similarity.reshape(1, -1) @ interaction_matrix).A1

    interacted_items = set(target_vector.indices.tolist())
    result: dict[int, float] = {}
    for product_id in candidate_product_ids:
        item_idx = item_to_idx.get(product_id)
        if item_idx is None:
            continue
        # 优先推荐“未交互”的商品；若全为空由上层兜底热门。
        if item_idx in interacted_items:
            continue
        score = float(predicted_scores[item_idx])
        if score > 0:
            result[product_id] = score

    return result


from __future__ import annotations

from typing import Iterable

try:
    import numpy as np
    from sklearn.feature_extraction.text import TfidfVectorizer
    from sklearn.metrics.pairwise import cosine_similarity
except Exception:  # pragma: no cover
    np = None
    TfidfVectorizer = None
    cosine_similarity = None


def compute_content_scores(
    user_item_scores: Iterable[dict],
    candidate_products: list[dict],
) -> dict[int, float]:
    if np is None or TfidfVectorizer is None or cosine_similarity is None:
        return {}
    if not candidate_products:
        return {}

    candidate_ids = [int(item["id"]) for item in candidate_products]
    id_to_index = {pid: idx for idx, pid in enumerate(candidate_ids)}

    texts: list[str] = []
    for item in candidate_products:
        name = str(item.get("name", "") or "")
        category = str(item.get("category", "") or "")
        texts.append(f"{name} {category}".strip())

    vectorizer = TfidfVectorizer(max_features=1200, ngram_range=(1, 2))
    item_matrix = vectorizer.fit_transform(texts)
    if item_matrix.shape[0] == 0:
        return {}

    interacted_indices: list[int] = []
    interacted_weights: list[float] = []
    for row in user_item_scores:
        product_id = int(row.get("product_id", 0) or 0)
        score = float(row.get("score", 0) or 0)
        idx = id_to_index.get(product_id)
        if idx is None or score <= 0:
            continue
        interacted_indices.append(idx)
        interacted_weights.append(score)

    if not interacted_indices:
        return {}

    weights = np.asarray(interacted_weights, dtype=float)
    if weights.sum() <= 0:
        return {}
    weights = weights / weights.sum()

    interacted_matrix = item_matrix[interacted_indices]
    profile_vector = (interacted_matrix.T @ weights.reshape(-1, 1)).T
    similarity_scores = cosine_similarity(profile_vector, item_matrix).flatten()

    interacted_set = set(interacted_indices)
    result: dict[int, float] = {}
    for idx, product_id in enumerate(candidate_ids):
        if idx in interacted_set:
            continue
        score = float(similarity_scores[idx])
        if score > 0:
            result[product_id] = score
    return result


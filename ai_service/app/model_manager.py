"""AI Model Manager — fetches model configs from backend and supports fallback switching.

This module provides:
1. Fetching available model list from the Go backend API
2. Selecting best model by priority
3. Fallback to next available model on failure
4. Usage reporting after each API call
"""

import logging
import time
from dataclasses import dataclass, field
from typing import Optional

import httpx

logger = logging.getLogger(__name__)


@dataclass
class ModelConfig:
    """A single AI model configuration fetched from backend."""
    id: int
    provider: str
    model_name: str
    api_key: str = ""
    base_url: str = ""
    price_per_k: float = 0.0
    priority: int = 0


@dataclass
class ModelManager:
    """Manages AI model selection, fallback, and usage reporting.

    Parameters:
        backend_base_url: e.g. "http://localhost:8081/api"
        timeout: HTTP timeout in seconds
    """
    backend_base_url: str = "http://localhost:8081/api"
    timeout: float = 10.0
    _cache: list[ModelConfig] = field(default_factory=list)
    _cache_time: float = 0.0
    _cache_ttl: float = 60.0  # refresh model list every 60 seconds

    def _http_client(self) -> httpx.Client:
        return httpx.Client(timeout=self.timeout, trust_env=False)

    def fetch_active_models(self, force: bool = False) -> list[ModelConfig]:
        """Fetch enabled models from backend. Uses cache if not expired."""
        now = time.time()
        if not force and self._cache and (now - self._cache_time) < self._cache_ttl:
            return self._cache

        try:
            with self._http_client() as client:
                resp = client.get(f"{self.backend_base_url}/ai-models/active")
                resp.raise_for_status()
                data = resp.json().get("data", [])
                self._cache = [
                    ModelConfig(
                        id=item.get("id") or item.get("ID"),
                        provider=item.get("provider", ""),
                        model_name=item.get("model_name", ""),
                        base_url=item.get("base_url", ""),
                        price_per_k=item.get("price_per_k", 0),
                        priority=item.get("priority", 0),
                    )
                    for item in data if (item.get("id") or item.get("ID"))
                ]
                self._cache_time = now
                logger.info(f"Fetched {len(self._cache)} active AI models from backend")
        except Exception as e:
            logger.warning(f"Failed to fetch AI models from backend: {e}")
            # Return cached data if available, even if stale
        return self._cache

    def fetch_model_secret(self, model_id: int) -> Optional[ModelConfig]:
        """Fetch full model config (with decrypted API key) from backend."""
        try:
            with self._http_client() as client:
                resp = client.get(f"{self.backend_base_url}/ai-models/secret/{model_id}")
                resp.raise_for_status()
                d = resp.json()
                return ModelConfig(
                    id=d.get("id") or d.get("ID"),
                    provider=d.get("provider", ""),
                    model_name=d.get("model_name", ""),
                    api_key=d.get("api_key", ""),
                    base_url=d.get("base_url", ""),
                )
        except Exception as e:
            logger.error(f"Failed to fetch model secret for id={model_id}: {e}")
            return None

    def select_best_model(self) -> Optional[ModelConfig]:
        """Select the highest-priority enabled model."""
        models = self.fetch_active_models()
        if not models:
            return None
        # Already sorted by priority desc from backend
        return models[0]

    def get_fallback_models(self, exclude_id: int) -> list[ModelConfig]:
        """Get fallback model list, excluding the failed one."""
        models = self.fetch_active_models()
        return [m for m in models if m.id != exclude_id]

    def report_usage(
        self,
        model_id: int,
        app_type: str,
        prompt_tokens: int = 0,
        output_tokens: int = 0,
        total_tokens: int = 0,
        session_id: str = "",
        user_id: int = 0,
    ) -> None:
        """Report token usage to backend for cost tracking."""
        payload = {
            "model_id": model_id,
            "app_type": app_type,
            "prompt_tokens": prompt_tokens,
            "output_tokens": output_tokens,
            "total_tokens": total_tokens or (prompt_tokens + output_tokens),
            "session_id": session_id,
            "user_id": user_id,
        }
        try:
            with self._http_client() as client:
                resp = client.post(f"{self.backend_base_url}/ai-models/usage", json=payload)
                resp.raise_for_status()
                logger.debug(f"Usage reported: model_id={model_id}, tokens={total_tokens}")
        except Exception as e:
            logger.warning(f"Failed to report usage: {e}")

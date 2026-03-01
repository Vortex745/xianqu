import os
from dataclasses import dataclass
from pathlib import Path

from dotenv import load_dotenv

# Always load env files from ai_service directory no matter current working directory.
_AI_SERVICE_DIR = Path(__file__).resolve().parents[2]
load_dotenv(_AI_SERVICE_DIR / ".env")
load_dotenv(_AI_SERVICE_DIR / ".env.local")

def _parse_float(name: str, default: float) -> float:
    raw = os.getenv(name, str(default)).strip()
    try:
        return float(raw)
    except ValueError as exc:
        raise RuntimeError(f"Invalid float env var: {name}={raw}") from exc


@dataclass(frozen=True)
class Settings:
    deepseek_api_key: str
    deepseek_base_url: str
    deepseek_model: str
    deepseek_temperature: float
    deepseek_timeout: float
    backend_api_base_url: str
    backend_timeout: float
    allowed_origins: list[str]

    @classmethod
    def from_env(cls) -> "Settings":
        # Optional: when empty, service can still run with dynamic model config
        # fetched from backend admin settings.
        deepseek_api_key = os.getenv("DEEPSEEK_API_KEY", "").strip()
        deepseek_base_url = os.getenv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1").strip()
        deepseek_model = os.getenv("DEEPSEEK_MODEL", "deepseek-chat").strip()
        deepseek_temperature = _parse_float("DEEPSEEK_TEMPERATURE", 0.4)
        deepseek_timeout = _parse_float("DEEPSEEK_TIMEOUT", 30)
        backend_api_base_url = os.getenv("BACKEND_API_BASE_URL", "http://localhost:8081/api").strip()
        backend_timeout = _parse_float("BACKEND_TIMEOUT", 12)
        allowed_origins = [
            origin.strip()
            for origin in os.getenv("ALLOWED_ORIGINS", "http://localhost:5173").split(",")
            if origin.strip()
        ]
        return cls(
            deepseek_api_key=deepseek_api_key,
            deepseek_base_url=deepseek_base_url,
            deepseek_model=deepseek_model,
            deepseek_temperature=deepseek_temperature,
            deepseek_timeout=deepseek_timeout,
            backend_api_base_url=backend_api_base_url,
            backend_timeout=backend_timeout,
            allowed_origins=allowed_origins,
        )

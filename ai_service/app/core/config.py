import os
import re
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


def _parse_allowed_origins(raw: str) -> list[str]:
    # Support comma/newline/semicolon separated origins from Vercel env UI.
    tokens = re.split(r"[,\n\r;]+", raw or "")
    origins: list[str] = []
    for token in tokens:
        origin = token.strip().strip('"').strip("'")
        if origin:
            origins.append(origin)
    return origins


def _parse_service_mode(raw: str, default: str = "guide") -> str:
    value = str(raw or default).strip().lower()
    if value in {"support", "conversion", "guide"}:
        return value
    return default


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
    agent_service_mode: str
    agent_brand_tone: str
    agent_telemetry_dir: str

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
        allowed_origins = _parse_allowed_origins(
            os.getenv("ALLOWED_ORIGINS", "http://localhost:5173")
        )
        agent_service_mode = _parse_service_mode(os.getenv("AGENT_SERVICE_MODE", "guide"))
        agent_brand_tone = os.getenv(
            "AGENT_BRAND_TONE",
            "像闲趣里靠谱的熟人摊主。先说结论，再给动作。短句，别空话。",
        ).strip()
        agent_telemetry_dir = os.getenv("AGENT_TELEMETRY_DIR", str((_AI_SERVICE_DIR / "logs" / "agent").resolve())).strip()
        return cls(
            deepseek_api_key=deepseek_api_key,
            deepseek_base_url=deepseek_base_url,
            deepseek_model=deepseek_model,
            deepseek_temperature=deepseek_temperature,
            deepseek_timeout=deepseek_timeout,
            backend_api_base_url=backend_api_base_url,
            backend_timeout=backend_timeout,
            allowed_origins=allowed_origins,
            agent_service_mode=agent_service_mode,
            agent_brand_tone=agent_brand_tone,
            agent_telemetry_dir=agent_telemetry_dir,
        )

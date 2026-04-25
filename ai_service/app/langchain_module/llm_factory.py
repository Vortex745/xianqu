import httpx
from langchain_openai import ChatOpenAI

from ..core.config import Settings
from ..model_manager import ModelConfig


# Default base URLs per provider
PROVIDER_BASE_URLS = {
    "deepseek": "https://api.deepseek.com/v1",
    "aliyun": "https://dashscope.aliyuncs.com/compatible-mode/v1",
    "minimax": "https://api.minimax.chat/v1",
    "zhipu": "https://open.bigmodel.cn/api/paas/v4",
}


def build_chat_llm(settings: Settings) -> ChatOpenAI:
    """Build a ChatOpenAI instance configured for DeepSeek (legacy config)."""

    # Use generous timeouts: connect=10s, read/write/pool=per-setting.
    timeout = httpx.Timeout(
        connect=10.0,
        read=settings.deepseek_timeout,
        write=settings.deepseek_timeout,
        pool=settings.deepseek_timeout,
    )

    # Ignore process-level proxy env vars to avoid accidental broken proxy
    # settings (for example 127.0.0.1:9) from breaking model connectivity.
    sync_client = httpx.Client(timeout=timeout, trust_env=False)
    async_client = httpx.AsyncClient(timeout=timeout, trust_env=False)

    return ChatOpenAI(
        model=settings.deepseek_model,
        api_key=settings.deepseek_api_key,
        base_url=settings.deepseek_base_url,
        temperature=settings.deepseek_temperature,
        timeout=settings.deepseek_timeout,
        http_client=sync_client,
        http_async_client=async_client,
    )


def build_chat_llm_from_config(
    model_config: ModelConfig,
    temperature: float = 0.4,
    timeout_seconds: float = 30.0,
) -> ChatOpenAI:
    """Build a ChatOpenAI instance from a dynamic ModelConfig (fetched from backend)."""

    base_url = model_config.base_url or PROVIDER_BASE_URLS.get(
        model_config.provider, "https://api.deepseek.com/v1"
    )
    
    # Auto-correct common DeepSeek misconfigurations
    model_name = model_config.model_name
    if model_config.provider == "deepseek":
        if model_name.strip().lower() == "deepseek":
            model_name = "deepseek-chat"
        if base_url and base_url.endswith("api.deepseek.com"):
            base_url = "https://api.deepseek.com/v1"

    timeout = httpx.Timeout(
        connect=10.0,
        read=timeout_seconds,
        write=timeout_seconds,
        pool=timeout_seconds,
    )

    sync_client = httpx.Client(timeout=timeout, trust_env=False)
    async_client = httpx.AsyncClient(timeout=timeout, trust_env=False)

    return ChatOpenAI(
        model=model_name,
        api_key=model_config.api_key,
        base_url=base_url,
        temperature=temperature,
        timeout=timeout_seconds,
        http_client=sync_client,
        http_async_client=async_client,
    )

import logging

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .agent_module.service import LangChainAgentService
from .api.routes import create_api_router
from .core.config import Settings
from .langchain_module.service import LangChainCustomerService
from .model_manager import ModelManager

logger = logging.getLogger(__name__)


def create_app() -> FastAPI:
    settings = Settings.from_env()
    logger.info("CORS allowed origins: %s", settings.allowed_origins)

    # Initialize model manager for dynamic model switching
    model_manager = ModelManager(
        backend_base_url=settings.backend_api_base_url,
        timeout=settings.backend_timeout,
    )

    assistant = LangChainCustomerService(settings, model_manager=model_manager)
    agent = LangChainAgentService(settings, model_manager=model_manager)

    app = FastAPI(title="xianqu-ai-service", version="1.0.0")
    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.allowed_origins,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    # Support both `/chat` and `/ai/chat` styles to avoid path mismatch
    # across different reverse-proxy setups.
    app.include_router(create_api_router(assistant, agent))
    app.include_router(create_api_router(assistant, agent), prefix="/ai")
    return app

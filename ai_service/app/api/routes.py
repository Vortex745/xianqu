from fastapi import APIRouter, HTTPException

from ..agent_module.service import LangChainAgentService
from ..langchain_module.service import LangChainCustomerService
from ..recommend_module.service import RecommenderService
from ..schemas.chat import (
    AgentChatRequest,
    ChatRequest,
    ChatResponse,
    RecommendRequest,
    RecommendResponse,
)


def _map_service_exception(prefix: str, exc: Exception) -> HTTPException:
    message = str(exc)
    if "DEEPSEEK_API_KEY is not configured" in message:
        return HTTPException(status_code=503, detail="ai service misconfigured: missing DEEPSEEK_API_KEY")
    if "Connection error" in message:
        return HTTPException(status_code=502, detail=f"{prefix} upstream connection failed")
    return HTTPException(status_code=500, detail=f"{prefix}: {message}")


def create_api_router(
    assistant: LangChainCustomerService,
    agent: LangChainAgentService,
    recommender: RecommenderService,
) -> APIRouter:
    router = APIRouter()

    @router.get("/health")
    def health() -> dict[str, str]:
        return {"status": "ok"}

    @router.post("/chat", response_model=ChatResponse)
    def chat(payload: ChatRequest) -> ChatResponse:
        try:
            session_id, answer = assistant.chat(payload.message, payload.session_id)
        except ValueError as exc:
            raise HTTPException(status_code=422, detail=str(exc)) from exc
        except Exception as exc:
            raise _map_service_exception("llm invoke failed", exc) from exc

        return ChatResponse(session_id=session_id, answer=answer)

    @router.post("/agent/chat", response_model=ChatResponse)
    def agent_chat(payload: AgentChatRequest) -> ChatResponse:
        try:
            session_id, answer, meta = agent.chat(
                message=payload.message,
                session_id=payload.session_id,
                auth_token=payload.auth_token,
                service_mode=payload.service_mode,
            )
        except ValueError as exc:
            raise HTTPException(status_code=422, detail=str(exc)) from exc
        except Exception as exc:
            raise _map_service_exception("agent invoke failed", exc) from exc

        return ChatResponse(session_id=session_id, answer=answer, meta=meta)

    @router.delete("/session/{session_id}")
    def clear_session(session_id: str) -> dict[str, bool]:
        return {"ok": True, "cleared": assistant.clear_session(session_id)}

    @router.delete("/agent/session/{session_id}")
    def clear_agent_session(session_id: str) -> dict[str, bool]:
        return {"ok": True, "cleared": agent.clear_session(session_id)}

    @router.post("/recommend", response_model=RecommendResponse)
    def recommend(payload: RecommendRequest) -> RecommendResponse:
        try:
            product_ids, source = recommender.recommend(
                user_id=payload.user_id,
                user_item_scores=payload.user_item_scores,
                behavior_rows=payload.behavior_rows,
                candidate_products=payload.candidate_products,
                top_k=payload.top_k,
            )
        except Exception as exc:
            raise HTTPException(status_code=500, detail=f"recommend failed: {exc}") from exc

        return RecommendResponse(product_ids=product_ids, source=source)

    return router

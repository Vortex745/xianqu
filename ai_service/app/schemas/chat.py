from pydantic import BaseModel, Field


class ChatRequest(BaseModel):
    message: str = Field(..., min_length=1, max_length=2000)
    session_id: str | None = Field(default=None, max_length=128)


class AgentChatRequest(BaseModel):
    message: str = Field(..., min_length=1, max_length=2000)
    session_id: str | None = Field(default=None, max_length=128)
    auth_token: str | None = Field(default=None, max_length=2048)
    service_mode: str | None = Field(default=None, max_length=32)


class QuickReply(BaseModel):
    label: str = Field(..., min_length=1, max_length=32)
    message: str = Field(..., min_length=1, max_length=200)


class ChatResponseMeta(BaseModel):
    service_mode: str | None = None
    source: str | None = None
    intent: str | None = None
    confidence: float | None = None
    sentiment: str | None = None
    fallback: bool = False
    clarification_kind: str | None = None
    guidance: str | None = None
    quick_replies: list[QuickReply] = Field(default_factory=list)


class ChatResponse(BaseModel):
    session_id: str
    answer: str
    meta: ChatResponseMeta | None = None

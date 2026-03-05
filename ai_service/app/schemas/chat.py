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


class UserItemScore(BaseModel):
    product_id: int = Field(..., ge=1)
    score: float = Field(..., ge=0)


class BehaviorRow(BaseModel):
    user_id: int = Field(..., ge=1)
    product_id: int = Field(..., ge=1)
    weight: float = Field(..., ge=0)


class CandidateProduct(BaseModel):
    id: int = Field(..., ge=1)
    name: str = ""
    category: str = ""
    price: float = 0
    view_count: int = 0
    popularity: float = 0


class RecommendRequest(BaseModel):
    user_id: int = 0
    user_item_scores: list[UserItemScore] = Field(default_factory=list)
    behavior_rows: list[BehaviorRow] = Field(default_factory=list)
    candidate_products: list[CandidateProduct] = Field(default_factory=list)
    top_k: int = 20


class RecommendResponse(BaseModel):
    product_ids: list[int] = Field(default_factory=list)
    source: str = "hot"

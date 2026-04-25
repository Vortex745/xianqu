import logging
from uuid import uuid4

from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from langchain_core.runnables.history import RunnableWithMessageHistory

from ..core.config import Settings
from ..model_manager import ModelManager
from .history_store import ChatHistoryStore
from .llm_factory import build_chat_llm, build_chat_llm_from_config
from .text_utils import as_text, normalize_plain_text

logger = logging.getLogger(__name__)


class LangChainCustomerService:
    def __init__(self, settings: Settings, history_store: ChatHistoryStore | None = None, model_manager: ModelManager | None = None) -> None:
        key = (settings.deepseek_api_key or "").strip()
        self._is_mock = not key or "replace_with_your_key" in key
        self._has_api_key = not self._is_mock
        self._settings = settings
        self._history_store = history_store or ChatHistoryStore()
        self._model_manager = model_manager
        if self._is_mock and not self._model_manager:
            self._chain = None
            return
        if not self._is_mock:
            llm = build_chat_llm(settings)
            self._chain = self._build_chain(llm)
        else:
            self._chain = None

    def _build_chain(self, llm):
        prompt = ChatPromptTemplate.from_messages(
            [
                (
                    "system",
                    "你是闲趣客服助手。"
                    "先说重点，再给步骤。"
                    "回答务实，别空话。"
                    "如果用户问退款、发货、账号问题，优先给可执行动作。"
                    "只输出纯文本，不要 Markdown、HTML、代码块。"
                    '需要分点时，请用"第一、第二、第三"表达。',
                ),
                MessagesPlaceholder(variable_name="history"),
                ("human", "{input}"),
            ]
        )
        return RunnableWithMessageHistory(
            prompt | llm,
            self._history_store.get_history,
            input_messages_key="input",
            history_messages_key="history",
        )

    def chat(self, message: str, session_id: str | None = None) -> tuple[str, str]:
        user_text = message.strip()
        if not user_text:
            raise ValueError("message is empty")

        use_session_id = session_id or uuid4().hex

        # Try dynamic model from backend if model_manager is available
        if self._model_manager:
            result = self._chat_with_dynamic_model(user_text, use_session_id)
            if result:
                return result

        if self._is_mock:
            mock_responses = {
                "你好": "你好！我是闲趣AI客服（演示模式）。由于暂未配置 API Key，我目前只能进行简单的复读回复。您可以在 ai_service/.env 中配置真实的 DEEPSEEK_API_KEY 以开启完整功能。",
                "hi": "Hello! I am in demo mode. Please configure DEEPSEEK_API_KEY to enable real AI features.",
            }
            answer = mock_responses.get(user_text, f"（演示模式）您说的是：{user_text}。 提示：请在 ai_service/.env 中配置真实的 API Key。")
            return use_session_id, answer

        if self._chain is None:
            return use_session_id, "AI 服务未初始化，请检查配置。"

        response = self._chain.invoke(
            {"input": user_text},
            config={"configurable": {"session_id": use_session_id}},
        )
        answer = normalize_plain_text(as_text(response.content))
        if not answer:
            answer = "我这边刚卡住了，请再试一次。"
        return use_session_id, answer

    def _chat_with_dynamic_model(self, user_text: str, session_id: str) -> tuple[str, str] | None:
        """Try to chat using dynamically configured models with fallback."""
        models = self._model_manager.fetch_active_models()
        if not models:
            return None

        tried_ids = set()
        for model_cfg in models:
            if model_cfg.id in tried_ids:
                continue
            tried_ids.add(model_cfg.id)

            secret = self._model_manager.fetch_model_secret(model_cfg.id)
            if not secret or not secret.api_key:
                logger.warning(f"Skipping model {model_cfg.id}: no API key")
                continue

            try:
                llm = build_chat_llm_from_config(
                    secret,
                    temperature=self._settings.deepseek_temperature,
                    timeout_seconds=self._settings.deepseek_timeout,
                )
                chain = self._build_chain(llm)
                response = chain.invoke(
                    {"input": user_text},
                    config={"configurable": {"session_id": session_id}},
                )
                answer = normalize_plain_text(as_text(response.content))
                if not answer:
                    answer = "我这边刚卡住了，请再试一次。"

                # Report usage
                usage = getattr(response, "response_metadata", {}).get("token_usage", {})
                self._model_manager.report_usage(
                    model_id=model_cfg.id,
                    app_type="customer_service",
                    prompt_tokens=usage.get("prompt_tokens", 0),
                    output_tokens=usage.get("completion_tokens", 0),
                    total_tokens=usage.get("total_tokens", 0),
                    session_id=session_id,
                )

                logger.info(f"Customer service used model: {secret.provider}/{secret.model_name}")
                return session_id, answer

            except Exception as e:
                logger.warning(f"Model {model_cfg.id} ({model_cfg.provider}/{model_cfg.model_name}) failed: {e}, trying fallback...")
                continue

        logger.warning("All dynamic models failed, falling back to static config")
        return None

    def clear_session(self, session_id: str) -> bool:
        return self._history_store.clear_session(session_id)

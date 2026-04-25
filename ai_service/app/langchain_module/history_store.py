from threading import Lock

from langchain_core.chat_history import BaseChatMessageHistory, InMemoryChatMessageHistory


class ChatHistoryStore:
    def __init__(self) -> None:
        self._store: dict[str, InMemoryChatMessageHistory] = {}
        self._lock = Lock()

    def get_history(self, session_id: str) -> BaseChatMessageHistory:
        with self._lock:
            if session_id not in self._store:
                self._store[session_id] = InMemoryChatMessageHistory()
            return self._store[session_id]

    def clear_session(self, session_id: str) -> bool:
        with self._lock:
            return self._store.pop(session_id, None) is not None

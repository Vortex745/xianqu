import re
from html import unescape


def as_text(content: object) -> str:
    if isinstance(content, str):
        return content.strip()
    if isinstance(content, list):
        parts: list[str] = []
        for block in content:
            if isinstance(block, dict):
                text = block.get("text")
                if isinstance(text, str) and text.strip():
                    parts.append(text.strip())
        return "\n".join(parts).strip()
    return str(content).strip()


def _replace_markdown_image(match: re.Match[str]) -> str:
    alt = (match.group(1) or "").strip()
    url = (match.group(2) or "").strip()
    label = alt if alt else "图片"
    return f"{label}（详见：{url}）" if url else label


def _replace_markdown_link(match: re.Match[str]) -> str:
    title = (match.group(1) or "").strip()
    url = (match.group(2) or "").strip()
    if title and url:
        return f"{title}（详见：{url}）"
    if url:
        return f"详见：{url}"
    return title


def normalize_plain_text(raw_text: str) -> str:
    text = unescape(str(raw_text or "")).replace("\r\n", "\n")

    # Keep code content but remove fenced markers.
    text = re.sub(r"```[^\n`]*\n?([\s\S]*?)```", lambda m: (m.group(1) or "").strip(), text)
    text = re.sub(r"`([^`]+)`", r"\1", text)

    # Convert markdown links/images to plain language.
    text = re.sub(r"!\[([^\]]*)\]\(([^)]+)\)", _replace_markdown_image, text)
    text = re.sub(r"\[([^\]]+)\]\(([^)]+)\)", _replace_markdown_link, text)

    # Strip HTML and heading markers.
    text = re.sub(r"<br\s*/?>", "\n", text, flags=re.IGNORECASE)
    text = re.sub(r"</p\s*>", "\n", text, flags=re.IGNORECASE)
    text = re.sub(r"<[^>]+>", "", text)

    normalized_lines: list[str] = []
    item_index = 0
    for raw_line in text.split("\n"):
        line = raw_line.strip()
        if not line:
            if normalized_lines and normalized_lines[-1] != "":
                normalized_lines.append("")
            item_index = 0
            continue

        line = re.sub(r"^#{1,6}\s*", "", line)
        line = re.sub(r"^>\s*", "", line)

        bullet = re.match(r"^[-*+]\s+(.*)", line)
        ordered = re.match(r"^\d+[.)、]\s+(.*)", line)
        if bullet:
            item_index += 1
            line = f"第{item_index}点：{bullet.group(1).strip()}"
        elif ordered:
            item_index += 1
            line = f"第{item_index}点：{ordered.group(1).strip()}"
        else:
            item_index = 0

        line = line.replace("**", "").replace("__", "").replace("~~", "").replace("*", "")
        line = line.replace("[", "").replace("]", "")
        line = re.sub(r"\s+", " ", line).strip()

        if line:
            normalized_lines.append(line)

    plain = "\n".join(normalized_lines).strip()
    plain = re.sub(r"\n{3,}", "\n\n", plain)
    return plain

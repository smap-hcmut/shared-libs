"""Helpers for extracting a public-facing URL from a UAP metadata dict.

Mirrors shared-libs/go/uapurl so Python services (analysis-srv) and Go
services agree on the candidate key order and the trim semantics.
"""

from __future__ import annotations

from typing import Any, Iterable, Mapping

CANDIDATE_KEYS: tuple[str, ...] = (
    "post_url",
    "url",
    "permalink_url",
    "original_url",
    "source_url",
    "web_url",
    "comment_url",
    "share_url",
    "parent_post_url",
)


def first_from_mapping(metadata: Mapping[str, Any] | None) -> str:
    """Return the first non-empty URL found in *metadata*.

    The check walks CANDIDATE_KEYS in order; any value that is not a string
    or is empty after trim is skipped. Returns ``""`` when nothing matches.
    """

    if not metadata:
        return ""
    for key in CANDIDATE_KEYS:
        value = metadata.get(key)
        if isinstance(value, str):
            trimmed = value.strip()
            if trimmed:
                return trimmed
    return ""


def iter_urls(metadata: Mapping[str, Any] | None) -> Iterable[str]:
    """Yield each non-empty URL in CANDIDATE_KEYS order."""

    if not metadata:
        return
    for key in CANDIDATE_KEYS:
        value = metadata.get(key)
        if isinstance(value, str):
            trimmed = value.strip()
            if trimmed:
                yield trimmed


__all__ = ["CANDIDATE_KEYS", "first_from_mapping", "iter_urls"]

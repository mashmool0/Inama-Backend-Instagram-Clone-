import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Post(_message.Message):
    __slots__ = ("id", "author_id", "caption", "media_url", "like_count", "comment_count", "created_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    AUTHOR_ID_FIELD_NUMBER: _ClassVar[int]
    CAPTION_FIELD_NUMBER: _ClassVar[int]
    MEDIA_URL_FIELD_NUMBER: _ClassVar[int]
    LIKE_COUNT_FIELD_NUMBER: _ClassVar[int]
    COMMENT_COUNT_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    id: str
    author_id: str
    caption: str
    media_url: str
    like_count: int
    comment_count: int
    created_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[str] = ..., author_id: _Optional[str] = ..., caption: _Optional[str] = ..., media_url: _Optional[str] = ..., like_count: _Optional[int] = ..., comment_count: _Optional[int] = ..., created_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class Comment(_message.Message):
    __slots__ = ("id", "post_id", "author_id", "body", "created_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    POST_ID_FIELD_NUMBER: _ClassVar[int]
    AUTHOR_ID_FIELD_NUMBER: _ClassVar[int]
    BODY_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    id: str
    post_id: str
    author_id: str
    body: str
    created_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[str] = ..., post_id: _Optional[str] = ..., author_id: _Optional[str] = ..., body: _Optional[str] = ..., created_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class CreatePostRequest(_message.Message):
    __slots__ = ("caption", "media_url")
    CAPTION_FIELD_NUMBER: _ClassVar[int]
    MEDIA_URL_FIELD_NUMBER: _ClassVar[int]
    caption: str
    media_url: str
    def __init__(self, caption: _Optional[str] = ..., media_url: _Optional[str] = ...) -> None: ...

class DeletePostRequest(_message.Message):
    __slots__ = ("post_id",)
    POST_ID_FIELD_NUMBER: _ClassVar[int]
    post_id: str
    def __init__(self, post_id: _Optional[str] = ...) -> None: ...

class DeletePostResponse(_message.Message):
    __slots__ = ("success",)
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    success: bool
    def __init__(self, success: _Optional[bool] = ...) -> None: ...

class GetPostRequest(_message.Message):
    __slots__ = ("post_id",)
    POST_ID_FIELD_NUMBER: _ClassVar[int]
    post_id: str
    def __init__(self, post_id: _Optional[str] = ...) -> None: ...

class LikePostRequest(_message.Message):
    __slots__ = ("post_id",)
    POST_ID_FIELD_NUMBER: _ClassVar[int]
    post_id: str
    def __init__(self, post_id: _Optional[str] = ...) -> None: ...

class UnlikePostRequest(_message.Message):
    __slots__ = ("post_id",)
    POST_ID_FIELD_NUMBER: _ClassVar[int]
    post_id: str
    def __init__(self, post_id: _Optional[str] = ...) -> None: ...

class LikeResponse(_message.Message):
    __slots__ = ("like_count",)
    LIKE_COUNT_FIELD_NUMBER: _ClassVar[int]
    like_count: int
    def __init__(self, like_count: _Optional[int] = ...) -> None: ...

class AddCommentRequest(_message.Message):
    __slots__ = ("post_id", "body")
    POST_ID_FIELD_NUMBER: _ClassVar[int]
    BODY_FIELD_NUMBER: _ClassVar[int]
    post_id: str
    body: str
    def __init__(self, post_id: _Optional[str] = ..., body: _Optional[str] = ...) -> None: ...

class GetCommentsRequest(_message.Message):
    __slots__ = ("post_id", "limit", "cursor")
    POST_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    CURSOR_FIELD_NUMBER: _ClassVar[int]
    post_id: str
    limit: int
    cursor: str
    def __init__(self, post_id: _Optional[str] = ..., limit: _Optional[int] = ..., cursor: _Optional[str] = ...) -> None: ...

class CommentPage(_message.Message):
    __slots__ = ("comments", "next_cursor")
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    NEXT_CURSOR_FIELD_NUMBER: _ClassVar[int]
    comments: _containers.RepeatedCompositeFieldContainer[Comment]
    next_cursor: str
    def __init__(self, comments: _Optional[_Iterable[_Union[Comment, _Mapping]]] = ..., next_cursor: _Optional[str] = ...) -> None: ...

from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SearchType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SEARCH_TYPE_UNSPECIFIED: _ClassVar[SearchType]
    SEARCH_TYPE_USER: _ClassVar[SearchType]
    SEARCH_TYPE_HASHTAG: _ClassVar[SearchType]
    SEARCH_TYPE_TEXT: _ClassVar[SearchType]
SEARCH_TYPE_UNSPECIFIED: SearchType
SEARCH_TYPE_USER: SearchType
SEARCH_TYPE_HASHTAG: SearchType
SEARCH_TYPE_TEXT: SearchType

class SearchRequest(_message.Message):
    __slots__ = ("query", "type", "limit", "cursor")
    QUERY_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    CURSOR_FIELD_NUMBER: _ClassVar[int]
    query: str
    type: SearchType
    limit: int
    cursor: str
    def __init__(self, query: _Optional[str] = ..., type: _Optional[_Union[SearchType, str]] = ..., limit: _Optional[int] = ..., cursor: _Optional[str] = ...) -> None: ...

class UserResult(_message.Message):
    __slots__ = ("user_id", "username", "avatar_url")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    USERNAME_FIELD_NUMBER: _ClassVar[int]
    AVATAR_URL_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    username: str
    avatar_url: str
    def __init__(self, user_id: _Optional[str] = ..., username: _Optional[str] = ..., avatar_url: _Optional[str] = ...) -> None: ...

class PostResult(_message.Message):
    __slots__ = ("post_id", "author_id", "caption", "media_url")
    POST_ID_FIELD_NUMBER: _ClassVar[int]
    AUTHOR_ID_FIELD_NUMBER: _ClassVar[int]
    CAPTION_FIELD_NUMBER: _ClassVar[int]
    MEDIA_URL_FIELD_NUMBER: _ClassVar[int]
    post_id: str
    author_id: str
    caption: str
    media_url: str
    def __init__(self, post_id: _Optional[str] = ..., author_id: _Optional[str] = ..., caption: _Optional[str] = ..., media_url: _Optional[str] = ...) -> None: ...

class SearchResponse(_message.Message):
    __slots__ = ("users", "posts", "next_cursor")
    USERS_FIELD_NUMBER: _ClassVar[int]
    POSTS_FIELD_NUMBER: _ClassVar[int]
    NEXT_CURSOR_FIELD_NUMBER: _ClassVar[int]
    users: _containers.RepeatedCompositeFieldContainer[UserResult]
    posts: _containers.RepeatedCompositeFieldContainer[PostResult]
    next_cursor: str
    def __init__(self, users: _Optional[_Iterable[_Union[UserResult, _Mapping]]] = ..., posts: _Optional[_Iterable[_Union[PostResult, _Mapping]]] = ..., next_cursor: _Optional[str] = ...) -> None: ...

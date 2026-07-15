import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Profile(_message.Message):
    __slots__ = ("id", "username", "bio", "avatar_url", "follower_count", "following_count", "created_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    USERNAME_FIELD_NUMBER: _ClassVar[int]
    BIO_FIELD_NUMBER: _ClassVar[int]
    AVATAR_URL_FIELD_NUMBER: _ClassVar[int]
    FOLLOWER_COUNT_FIELD_NUMBER: _ClassVar[int]
    FOLLOWING_COUNT_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    id: str
    username: str
    bio: str
    avatar_url: str
    follower_count: int
    following_count: int
    created_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[str] = ..., username: _Optional[str] = ..., bio: _Optional[str] = ..., avatar_url: _Optional[str] = ..., follower_count: _Optional[int] = ..., following_count: _Optional[int] = ..., created_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class GetProfileRequest(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class UpdateProfileRequest(_message.Message):
    __slots__ = ("username", "bio", "avatar_url")
    USERNAME_FIELD_NUMBER: _ClassVar[int]
    BIO_FIELD_NUMBER: _ClassVar[int]
    AVATAR_URL_FIELD_NUMBER: _ClassVar[int]
    username: str
    bio: str
    avatar_url: str
    def __init__(self, username: _Optional[str] = ..., bio: _Optional[str] = ..., avatar_url: _Optional[str] = ...) -> None: ...

class FollowRequest(_message.Message):
    __slots__ = ("target_user_id",)
    TARGET_USER_ID_FIELD_NUMBER: _ClassVar[int]
    target_user_id: str
    def __init__(self, target_user_id: _Optional[str] = ...) -> None: ...

class FollowResponse(_message.Message):
    __slots__ = ("success",)
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    success: bool
    def __init__(self, success: _Optional[bool] = ...) -> None: ...

class UnfollowRequest(_message.Message):
    __slots__ = ("target_user_id",)
    TARGET_USER_ID_FIELD_NUMBER: _ClassVar[int]
    target_user_id: str
    def __init__(self, target_user_id: _Optional[str] = ...) -> None: ...

class UnfollowResponse(_message.Message):
    __slots__ = ("success",)
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    success: bool
    def __init__(self, success: _Optional[bool] = ...) -> None: ...

class GetFollowersRequest(_message.Message):
    __slots__ = ("user_id", "limit", "cursor")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    CURSOR_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    limit: int
    cursor: str
    def __init__(self, user_id: _Optional[str] = ..., limit: _Optional[int] = ..., cursor: _Optional[str] = ...) -> None: ...

class GetFollowingRequest(_message.Message):
    __slots__ = ("user_id", "limit", "cursor")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    CURSOR_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    limit: int
    cursor: str
    def __init__(self, user_id: _Optional[str] = ..., limit: _Optional[int] = ..., cursor: _Optional[str] = ...) -> None: ...

class UserIdPage(_message.Message):
    __slots__ = ("user_ids", "next_cursor")
    USER_IDS_FIELD_NUMBER: _ClassVar[int]
    NEXT_CURSOR_FIELD_NUMBER: _ClassVar[int]
    user_ids: _containers.RepeatedScalarFieldContainer[str]
    next_cursor: str
    def __init__(self, user_ids: _Optional[_Iterable[str]] = ..., next_cursor: _Optional[str] = ...) -> None: ...

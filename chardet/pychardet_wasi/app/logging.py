"""Minimal logging shim for the embedded CPython WASI runtime.

chardet only emits DEBUG-level diagnostics.  The full logging package pulls in
traceback and a large part of the standard library, which the reduced runtime
does not ship.
"""

DEBUG = 10


class _Logger:
    def getEffectiveLevel(self) -> int:
        return 0

    def debug(self, *args, **kwargs) -> None:
        pass


def getLogger(name=None) -> _Logger:
    return _Logger()

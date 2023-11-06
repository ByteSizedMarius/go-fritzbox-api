import os
from datetime import datetime


def expect_args(args, n):
    if len(args) != n:
        out(f"Invalid arguments. Expected {n - 1} arguments, got {len(args) - 1}.")
        return False
    return True


def out(msg):
    print(msg, flush=True)


def ok():
    out("OK")


def to_html(text):
    import tempfile
    with open(os.path.join(tempfile.gettempdir(), f"source{datetime.now().microsecond}.html"), "w") as f:
        f.write(text)


def urljoin(url, join):
    if url.endswith("/"):
        return url + join
    else:
        return url + "/" + join

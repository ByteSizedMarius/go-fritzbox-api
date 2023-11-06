import sys
import traceback

from dev.pyadapter import PyAdapter
from dev.util import out

if __name__ == '__main__':
    out("HELO")

    py = PyAdapter()
    if len(sys.argv) > 1:
        py.do(inp=" ".join(sys.argv[1:]).split(";"))
        exit(0)

    first_inp = sys.stdin.readline()
    if first_inp == "OK\n":
        # noinspection PyBroadException
        try:
            py.do()
        except:
            out("Error: " + traceback.format_exc().replace("\n", "//"))
    else:
        print("Invalid OK: " + repr(first_inp))

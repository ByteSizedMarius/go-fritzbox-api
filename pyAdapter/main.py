import datetime

try:
    import json
    from selenium.common import TimeoutException, NoSuchElementException
    from seleniumwire import webdriver
    import sys
    from selenium.webdriver.common.by import By
    from selenium.webdriver.support.wait import WebDriverWait
    from selenium.webdriver.support import expected_conditions as ec
    import traceback
except (ImportError, ModuleNotFoundError) as e:
    print("Error: " + repr(e), flush=True)
    exit()

params_exclude = ["xhr", "sid", "view", "lang"]


def expect_args(args, n):
    if len(args) != n:
        out("Invalid arguments. Expected " + str(n - 1) + " arguments, got " + str(len(args) - 1) + ".")
        return False
    return True


browser: webdriver.Chrome = None


def main(inp=None):
    if inp is None:
        inp = sys.stdin

    debug = False
    headless = True
    driverargs = []

    for line in inp:
        args = line.split()

        match args[0].lower():
            case "browser_debug":
                if not expect_args(args, 1): continue
                headless = False
                out("OK")
            case "debug":
                if not expect_args(args, 1): continue
                debug = True
                out("OK")
            case "args":
                if not expect_args(args, 2): continue
                driverargs = args[1].split("|")
                out("OK")
            case "login":
                if not expect_args(args, 3): continue
                login(headless, debug, driverargs, *args[1:])
            case "hkr":
                if not expect_args(args, 3): continue
                hkr(debug, *args[1:])
            case _:
                out("Invalid command: " + args[0])


def login(headless, debug, args, url, sid):
    global browser

    # start browser initially
    if not browser:
        options = webdriver.ChromeOptions()
        options.add_experimental_option("excludeSwitches", ["enable-logging"])

        if headless:
            options.add_argument("--headless")

        for arg in args:
            options.add_argument(arg)

        browser = webdriver.Chrome(options=options)

    # go to the login page
    browser.get(f"{url}/?sid={sid}")

    try:
        browser.find_element(By.ID, "uiPassInput")
        out("Invalid sid")
        exit(0)

    except NoSuchElementException:
        ...

    # check if we're on the homepage
    if not expect((By.ID, "blueBarUserMenuIcon")):
        out("Unknown Error while logging in")
        # write source to file, so we can see where we are
        if debug: to_html()
        return

    out("OK")


def to_html():
    with open(f"source{datetime.datetime.now().microsecond}.html", "w") as f:
        f.write(browser.page_source)


def urljoin(url, join):
    if url.endswith("/"):
        return url + join
    else:
        return url + "/" + join


def expect(elem):
    try:
        WebDriverWait(browser, 5).until(ec.presence_of_element_located(elem))
        return True
    except TimeoutException:
        return False


def hkr(debug, url, dev_id):
    browser.get(urljoin(url, "#sh_dev"))

    smarthome_table_find = (By.CLASS_NAME, "smarthome-devices")
    if not expect(smarthome_table_find):
        out("Could not find table smarthome-devices")
        if debug: to_html()
        return
    smarthome_table = browser.find_element(*smarthome_table_find)

    # click any edit button, intercept the request and redirect it to the correct device
    def intercept_hkr(request):
        if request.method == "POST" and "page" in request.params and request.params["page"] == "home_auto_hkr_edit":

            # request to go to edit page
            if "back_to_page" not in request.params:
                params = request.params
                params["device"] = dev_id
                request.params = params

            # apply request
            elif request.params["back_to_page"] == "/smarthome/devices.lua":
                out("SUCCESS " + json.dumps({k: v for k, v in request.params.items() if k not in params_exclude}))

    browser.request_interceptor = intercept_hkr

    edit_btn_find = (By.CLASS_NAME, "v-icon--edit")
    if not expect(edit_btn_find):
        out("Could not find table edit button")
        return

    edit_btn = smarthome_table.find_element(*edit_btn_find)
    edit_btn.find_element(By.XPATH, "../..").click()

    # wait for the edit page
    if not expect((By.NAME, "mainform")):
        out("Could not load HKR edit page")
        return
    if not expect((By.ID, "uiTimerArea")):
        out("Could not load HKR edit page")
        return

    # the apply button has the html disabled attribute unless you change something
    # it's only visual, so we can just remove it
    apply_button_find = (By.ID, "uiMainApply")
    apply_button = browser.find_element(*apply_button_find)
    if not apply_button:
        out("Could not find apply button")
        return

    browser.execute_script("arguments[0].removeAttribute('disabled')", apply_button)
    try:
        WebDriverWait(browser, 5).until(ec.element_to_be_clickable(apply_button_find))
    except TimeoutException:
        out("Apply button not clickable")
        return

    apply_button.click()
    browser.get(urljoin(url, "#overview"))


def out(msg):
    print(msg, flush=True)


if __name__ == '__main__':
    out("HELO")

    if len(sys.argv) > 1:
        main(inp=" ".join(sys.argv[1:]).split(";"))
        exit(0)

    first_inp = sys.stdin.readline()

    if first_inp == "OK\n":
        try:
            main()
        except:
            out("Error: " + traceback.format_exc().replace("\n", "//"))
    else:
        print("Invalid OK: " + repr(first_inp))

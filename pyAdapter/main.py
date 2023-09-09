import json
from selenium.common import TimeoutException, NoSuchElementException
from seleniumwire import webdriver
import sys
from selenium.webdriver.common.by import By
from selenium.webdriver.support.wait import WebDriverWait
from selenium.webdriver.support import expected_conditions as ec

params_exclude = ["xhr", "sid", "view", "lang"]


def expect_args(args, n):
    if len(args) != n:
        out("Invalid arguments")
        return False
    return True


def main():
    debug = False
    browser = None

    for line in sys.stdin:
        args = line.split()

        match args[0].lower():
            case "debug":
                if not expect_args(args, 1): continue
                debug = True
                out("OK")
            case "login":
                if not expect_args(args, 3): continue
                browser = login(debug, *args[1:])
            case "hkr":
                if not expect_args(args, 3): continue
                hkr(browser, *args[1:])
            case _:
                out("Invalid command")


def login(debug, url, sid):
    options = webdriver.ChromeOptions()
    options.add_experimental_option("excludeSwitches", ["enable-logging"])

    if not debug:
        options.add_argument("--headless")

    browser = webdriver.Chrome(options=options)
    browser.get(f"{url}/?sid={sid}")

    try:
        browser.find_element(By.ID, "uiPassInput")
        out("Invalid sid")
        return
    except NoSuchElementException:
        ...

    out("OK")
    return browser


def urljoin(url, join):
    if url.endswith("/"):
        return url + join
    else:
        return url + "/" + join


def expect(browser, elem):
    try:
        WebDriverWait(browser, 5).until(ec.presence_of_element_located(elem))
        return True
    except TimeoutException:
        return False


def hkr(browser, url, dev_id):
    browser.get(urljoin(url, "#sh_dev"))

    smarthome_table_find = (By.CLASS_NAME, "smarthome-devices")
    if not expect(browser, smarthome_table_find):
        out("Could not find table smarthome-devices")
        return
    smarthome_table = browser.find_element(*smarthome_table_find)

    # click any edit button, intercept the request and redirect it to the correct device
    def intercept_hkr(request):
        if request.method == 'POST' and "page" in request.params and request.params["page"] == "home_auto_hkr_edit":

            # request to go to edit page
            if "back_to_page" not in request.params:
                params = request.params
                params["device"] = dev_id
                request.params = params

            # apply request
            elif request.params["back_to_page"] == "/smarthome/devices.lua":
                out("SUCCESS " + json.dumps({k: v for k, v in request.params.items() if k not in params_exclude}))

    browser.request_interceptor = intercept_hkr
    edit_btn = smarthome_table.find_element(By.CLASS_NAME, "v-icon--edit")
    if not edit_btn:
        out("Could not find edit button")
        return
    edit_btn.find_element(By.XPATH, "../..").click()

    # wait for the edit page
    form_find = (By.NAME, "mainform")
    if not expect(browser, form_find):
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

    inp = sys.stdin.readline()
    if inp == "OK\n":
        try:
            main()
        except Exception as e:
            out("Error: " + repr(e))
    else:
        print("Invalid OK: " + repr(inp))

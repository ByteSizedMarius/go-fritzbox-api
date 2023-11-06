import json
import sys
from threading import Lock

from util import expect_args, out, ok, to_html, urljoin

try:
    from selenium.common import TimeoutException, NoSuchElementException
    from seleniumwire import webdriver
    from selenium.webdriver.common.by import By
    from selenium.webdriver.support.wait import WebDriverWait
    from selenium.webdriver.support import expected_conditions as ec
except (ImportError, ModuleNotFoundError) as e:
    print("Error: " + repr(e), flush=True)
    exit()


class PyAdapter:
    # request-parameters not relevant for us
    params_exclude = ["xhr", "sid", "view", "lang"]

    def __init__(self):
        # Internal vars
        self.browser = None
        self.browser_mutex = Lock()

        # Settable vars
        self.base_url = ""
        self.debug = False
        self.headless = True
        self.driverargs = []

    def do(self, inp=None):
        if inp is None:
            inp = sys.stdin

        for line in inp:
            args = line.split()
            with self.browser_mutex:
                self.execute(*args)

    def execute(self, *args):
        match args[0].lower():
            case "browser_debug":
                if not expect_args(args, 1): return
                self.headless = False
                ok()
            case "debug":
                if not expect_args(args, 1): return
                self.debug = True
                ok()
            case "args":
                if not expect_args(args, 2): return
                self.driverargs = args[1].split("|")
                ok()
            case "login":
                if not expect_args(args, 3): return
                self.login(*args[1:])
            case "refresh":
                if not expect_args(args, 1): return
                self.refresh()
            case "hkr":
                if not expect_args(args, 2): return
                self.hkr(*args[1:])
            case _:
                out("Invalid command: " + args[0])

    def login(self, url, sid):
        self.base_url = url

        # start browser initially
        if not self.browser:
            options = webdriver.ChromeOptions()
            options.add_experimental_option("excludeSwitches", ["enable-logging"])

            if self.headless:
                options.add_argument("--headless")

            for arg in self.driverargs:
                options.add_argument(arg)

            self.browser = webdriver.Chrome(options=options)

        # go to the login page
        self.browser.get(f"{url}/?sid={sid}")

        if self.on_login_page():
            out("Invalid sid")
            exit(0)

        # check if we're on the homepage
        if not self.expect((By.ID, "blueBarUserMenuIcon")):
            out("Unknown Error while logging in")
            # write source to file, so we can see where we are
            if self.debug: to_html(self.browser.page_source)
            return

        out("OK")

    def refresh(self):
        self.goto("#sh_organize")
        if not self.expect((By.CLASS_NAME, "smarthome-organize")):
            out("logged out")
            return

        self.goto_sh()
        if not self.expect((By.CLASS_NAME, "smarthome-devices")):
            out("logged out")
            return

        ok()

    def hkr(self, dev_id):
        self.goto_sh()

        # wait at most 5 seconds for the page to load
        smarthome_table_find = (By.CLASS_NAME, "smarthome-devices")
        if not self.expect(smarthome_table_find):

            # check if we're logged out
            if self.on_login_page():
                out("Invalid or expired sid")
                exit(0)

            # otherwise give a generic error and allow debugging if enabled
            out("Could not find table smarthome-devices")
            if self.debug: to_html(self.browser.page_source)
            return
        smarthome_table = self.browser.find_element(*smarthome_table_find)

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
                    out("SUCCESS " + json.dumps({k: v for k, v in request.params.items() if k not in self.params_exclude}))

        self.browser.request_interceptor = intercept_hkr

        edit_btn_find = (By.CLASS_NAME, "v-icon--edit")
        if not self.expect(edit_btn_find):
            out("Could not find table edit button")
            return

        edit_btn = smarthome_table.find_element(*edit_btn_find)
        edit_btn.find_element(By.XPATH, "../..").click()

        # wait for the edit page
        if not self.expect((By.NAME, "mainform")):
            out("Could not load HKR edit page")
            return
        # it seems the timer area loads after the rest of the page, so wait for that too
        if not self.expect((By.ID, "uiTimerArea")):
            out("Could not load HKR edit page")
            return

        # the apply button has the html disabled attribute unless you change something
        # it's only visual, so we can just remove it
        apply_button_find = (By.ID, "uiMainApply")
        apply_button = self.browser.find_element(*apply_button_find)
        if not apply_button:
            out("Could not find apply button")
            return
        self.browser.execute_script("arguments[0].removeAttribute('disabled')", apply_button)
        try:
            WebDriverWait(self.browser, 5).until(ec.element_to_be_clickable(apply_button_find))
        except TimeoutException:
            out("Apply button not clickable")
            return
        apply_button.click()
        self.goto_sh()

    def on_login_page(self):
        return self.expect((By.ID, "uiPassInput"), short=True)

    def goto_sh(self):
        self.goto("#sh_dev")

    def goto(self, shortcode):
        # it's very important that we only change pages if we're not already there
        # if we're already on the page and load it again, we will be logged out (probably a bug in the firmware?)
        if self.browser.current_url.endswith(shortcode):
            return

        self.browser.get(urljoin(self.base_url, shortcode))

    def expect(self, elem, short=False):
        try:
            WebDriverWait(self.browser, 5 if not short else 1).until(ec.presence_of_element_located(elem))
            return True
        except TimeoutException:
            return False

package selenium_test

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

//	To run this test:
//	go test -test.run=GithubLogin$
// OR
//  go test -test.run=TestGithubLogin$ github.com/tebeka/selenium

const (
	TAKE_SCREENSHOT = false
	WANT_MESSAGE    = "Hi,"
	USERNAME        = ""
	PASSWORD        = ""
	TRAVIS_RUN      = false
)

// Saves a image taken by the webdriver, and saves it to the current
// folder.
func saveScreenshot(t *testing.T, wd selenium.WebDriver, name string) {
	t.Helper()
	if !TAKE_SCREENSHOT {
		return
	}
	screenshot, err := wd.Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	img, _, _ := image.Decode(bytes.NewReader(screenshot))
	out, err := os.Create("./" + name + ".png")
	if err != nil {
		t.Fatal(err)
	}

	err = png.Encode(out, img)
	if err != nil {
		t.Fatal(err)
	}
}

func sleep() {
	time.Sleep(time.Millisecond * 1000)
}

func TestGithubLogin(t *testing.T) {
	// Start a Selenium WebDriver server instance (if one is not already
	// running).
	if !TRAVIS_RUN {
		t.Skipf("Test must be skipped on travis")
	}
	const (
		seleniumPath    = "./drivers/selenium-server-standalone-3.141.59.jar"
		geckoDriverPath = "./drivers/geckodriver"
		port            = 8080
	)
	//TODO(meling) the selenium dependency in vendor seems to be broken; fix and reactivate these commented lines
	// opts := []selenium.ServiceOption{
	// 	selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
	// 	selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
	// 	selenium.Output(os.Stderr),            // Output debug information to STDERR.
	// }
	selenium.SetDebug(true)
	// service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox", "acceptSslCerts": true}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		t.Fatal(err)
	}
	defer wd.Quit()

	// SETUP DONE
	// Navigate to the index page.
	if err := wd.Get("https://itest.run"); err != nil {
		t.Fatalf("failed on Get(\"https://www.itest.run\"): %v", err)
	}
	// Need time to load the index screen.
	sleep()

	// One way to get image of how the page you are currently at
	// looks like.
	saveScreenshot(t, wd, "indexpage")

	// It is also possible to insert script like below, this script does the exact same as
	// the css selector does.
	//wd.ExecuteScript("document.getElementsByClassName(\"social-login\")[0].children[0].firstChild.click()", nil)
	login, err := wd.FindElement(selenium.ByCSSSelector, "a[href=\"/app/login/login/github\"]")
	if err != nil {
		t.Fatal(err)
	}
	login.Click()
	sleep()

	// Fetch the fields needed.
	loginField, err := wd.FindElement(selenium.ByID, "login_field")
	if err != nil {
		t.Fatal(err)
	}

	loginField.SendKeys(USERNAME)

	passwordField, err := wd.FindElement(selenium.ByID, "password")
	if err != nil {
		t.Fatal(err)
	}
	passwordField.SendKeys(PASSWORD)

	loginButton, err := wd.FindElement(selenium.ByName, "commit")
	if err != nil {
		t.Fatal(err)
	}

	// One way to get image of how the page you are currently at
	// looks like.
	saveScreenshot(t, wd, "loginpage")

	loginButton.Click()
	sleep()

	// One way to get image of how the page you are currently at
	// looks like.
	saveScreenshot(t, wd, "AuthPage")

	authorize, err := wd.FindElement(selenium.ByID, "js-oauth-authorize-btn")
	if err != nil {
		t.Errorf("no auth page found: %v", err)
	} else {
		authorize.Click()
	}

	sleep()
	hellomsg, err := wd.FindElement(selenium.ByCSSSelector, "div[class='centerblock container']")
	if err != nil {
		t.Fatal(err)
	}

	outputText, err := hellomsg.Text()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(outputText, WANT_MESSAGE) {
		t.Errorf("have database course %+v want %+v", outputText, WANT_MESSAGE)
	}
}

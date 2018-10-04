package handler

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
)

var store = sessions.NewCookieStore([]byte("salada"))

const sessionName = "authentication"

type Worker struct {
	Client hydra.SDK
}

type User struct {
	Name     string
	Password string
}

func (w Worker) HandlerConsent(c echo.Context) error {

	consentRequestID := c.QueryParam("consent")
	if consentRequestID == "" {
		return c.JSON(http.StatusBadRequest, "Consent endpoint was called without a consent request id")

	}
	consentRequest, response, err := w.Client.GetOAuth2ConsentRequest(consentRequestID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "The consent request endpoint does not respond")

	} else if response.StatusCode != http.StatusOK {
		return c.JSON(http.StatusBadRequest, "Consent request endpoint")

	}

	request := c.Request()
	user := authenticated(request)
	if user == "" {
		return c.Redirect(http.StatusMovedPermanently, "/login?consent="+consentRequestID)

	}
	var s []string
	s = append(s, "authorization_code")
	response, err = w.Client.AcceptOAuth2ConsentRequest(consentRequestID, swagger.ConsentRequestAcceptance{
		Subject:     user,
		GrantScopes: s,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, "The accept consent request endpoint encountered a network error")
	} else if response.StatusCode != http.StatusNoContent {
		return c.JSON(http.StatusBadRequest, "ERRROR")

	}

	return c.Redirect(http.StatusMovedPermanently, consentRequest.RedirectUrl)
}

func (w Worker) HandlerLogin(c echo.Context) error {

	consentRequestID := c.QueryParam("consent")
	if consentRequestID == "" {
		return c.JSON(http.StatusBadRequest, "Consent endpoint was called without a consent request id")

	}
	recv := &User{
		Name:     "user_id",
		Password: "user_password",
	}
	if recv.Name != "user_id" || recv.Password != "user_password" {
		return c.JSON(http.StatusBadRequest, "User or Password incorrect")
	}

	request := c.Request()
	response := c.Response()
	session, _ := store.Get(request, sessionName)
	session.Values["user"] = recv.Name + recv.Password
	fmt.Println(session.Values["user"])
	if err := store.Save(request, response.Writer, session); err != nil {
		return c.JSON(http.StatusBadRequest, "error to save section")

	}

	return c.Redirect(http.StatusMovedPermanently, "/consent?consent="+consentRequestID)
}

func authenticated(r *http.Request) string {
	session, _ := store.Get(r, sessionName)
	if u, ok := session.Values["user"]; !ok {
		return ""
	} else if user, ok := u.(string); !ok {
		return ""
	} else {
		return user
	}
}

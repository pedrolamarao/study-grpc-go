package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type DeviceCode struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationUri         string `json:"verification_uri"`
	VerificationUriComplete string `json:"verification_uri_complete"`
	ExpiresIn               uint   `json:"expires_in"`
	Interval                uint   `json:"interval"`
}

func requestDeviceCode(client *http.Client, scope string) (*DeviceCode, error) {
	audience := os.Getenv("AUTH0_AUDIENCE")
	clientId := os.Getenv("AUTH0_CLIENT_ID")
	domain := os.Getenv("AUTH0_DOMAIN")
	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprint("https://", domain, "/oauth/device/code"),
		strings.NewReader(
			fmt.Sprint(
				"audience=", audience,
				"&scope=", scope,
				"&client_id=", clientId,
			),
		),
	)
	if err != nil {
		return nil, err
	}
	request.Header.Add("accept", "application/json")
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	deviceCode := DeviceCode{}
	err = json.Unmarshal(body, &deviceCode)
	if err != nil {
		return nil, err
	}
	return &deviceCode, nil
}

type AccessToken struct {
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    uint   `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type AccessTokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

var (
	accessTokenAuthorizationPending = errors.New("authorization_pending")
	accessTokenSlowDown             = errors.New("slow_down")
	accessTokenAccessDenied         = errors.New("access_denied")
)

func requestAccessToken(client *http.Client, scope string, deviceCode string) (*AccessToken, error) {
	audience := os.Getenv("AUTH0_AUDIENCE")
	clientId := os.Getenv("AUTH0_CLIENT_ID")
	domain := os.Getenv("AUTH0_DOMAIN")
	request, err := http.NewRequest(
		"POST",
		fmt.Sprint("https://", domain, "/oauth/token"),
		strings.NewReader(
			fmt.Sprint(
				"grant_type=urn:ietf:params:oauth:grant-type:device_code",
				"&audience=", audience,
				"&scope=", scope,
				"&client_id=", clientId,
				"&device_code=", deviceCode,
			),
		),
	)
	if err != nil {
		return nil, err
	}
	request.Header.Add("accept", "application/json")
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == 200 {
		accessToken := AccessToken{}
		err = json.Unmarshal(body, &accessToken)
		if err != nil {
			return nil, err
		}
		return &accessToken, nil
	}
	accessTokenError := AccessTokenError{}
	err = json.Unmarshal(body, &accessTokenError)
	if err != nil {
		return nil, err
	}
	switch accessTokenError.Error {
	case "authorization_pending":
		return nil, accessTokenAuthorizationPending
	case "slow_down":
		return nil, accessTokenSlowDown
	case "access_denied":
		return nil, accessTokenAccessDenied
	default:
		return nil, errors.New(accessTokenError.Error)
	}
}

func RequestAuthorization(client *http.Client, scope string) (string, error) {
	deviceCode, err := requestDeviceCode(client, scope)
	if err != nil {
		return "", err
	}
	fmt.Println("login: ", deviceCode.VerificationUriComplete)
	var accessToken *AccessToken
	for {
		time.Sleep(time.Duration(deviceCode.Interval) * time.Second)
		accessToken, err = requestAccessToken(client, scope, deviceCode.DeviceCode)
		if err == nil {
			break
		}
		if errors.Is(err, accessTokenAuthorizationPending) {
			continue
		}
		return "", err
	}
	return accessToken.AccessToken, nil
}

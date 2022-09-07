package auth

import (
	"errors"
	"fmt"
	"github.com/Levan-D/Todo-Backend/pkg/config"
	"github.com/Levan-D/Todo-Backend/pkg/storage"
	"github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
	"strings"
	"time"
)

type AccessDetails struct {
	AccessUUID string
	UserID     uuid.UUID
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

func CreateToken(userId uuid.UUID) (*TokenDetails, error) {
	td := &TokenDetails{}

	td.AtExpires = time.Now().Add(config.Get().JWT.AccessTTL).Unix()
	td.AccessUUID = fmt.Sprintf("%s", uuid.NewV4())

	td.RtExpires = time.Now().Add(config.Get().JWT.RefreshTTL).Unix()
	td.RefreshUUID = fmt.Sprintf("%s++%s", td.AccessUUID, userId)

	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUUID
	atClaims["user_id"] = userId
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(config.Get().JWT.AccessSecret))
	if err != nil {
		return nil, err
	}
	//Creating Refresh Token
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUUID
	rtClaims["user_id"] = userId
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(config.Get().JWT.RefreshSecret))
	if err != nil {
		return nil, err
	}
	return td, nil
}

func CreateAuth(userId uuid.UUID, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errAccess := storage.Set(td.AccessUUID, userId, at.Sub(now))
	if errAccess != nil {
		return errAccess
	}
	errRefresh := storage.Set(td.RefreshUUID, userId, rt.Sub(now))
	if errRefresh != nil {
		return errRefresh
	}
	return nil
}

func ExtractToken(r *fasthttp.Request) string {
	bearToken := r.Header.Peek("Authorization")
	strArr := strings.Split(string(bearToken), " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func VerifyToken(r *fasthttp.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Get().JWT.AccessSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func TokenValid(r *fasthttp.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok || !token.Valid {
		return err
	}
	return nil
}

func ExtractTokenMetadata(r *fasthttp.Request) (*AccessDetails, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid := claims["access_uuid"].(string)
		userId, _ := uuid.FromString(claims["user_id"].(string))
		return &AccessDetails{
			AccessUUID: accessUuid,
			UserID:     userId,
		}, nil
	}
	return nil, err
}

func FetchAuth(authD *AccessDetails) (string, error) {
	userid, err := storage.Get(authD.AccessUUID)
	if err != nil {
		return "", err
	}
	if authD.UserID.String() != userid {
		return "", errors.New("unauthorized")
	}
	return userid, nil
}

func DeleteAuth(givenUuid string) (int64, error) {
	deleted, err := storage.Delete(givenUuid)
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

func DeleteTokens(authD *AccessDetails) error {
	//get the refresh uuid
	refreshUuid := fmt.Sprintf("%s++%s", authD.AccessUUID, authD.UserID)
	//delete access token
	deletedAt, err := storage.Delete(authD.AccessUUID)
	if err != nil {
		return err
	}
	//delete refresh token
	deletedRt, err := storage.Delete(refreshUuid)
	if err != nil {
		return err
	}
	//When the record is deleted, the return value is 1
	if deletedAt != 1 || deletedRt != 1 {
		return errors.New("something went wrong")
	}
	return nil
}

func DeleteAllTokensByUserID(authD *AccessDetails) error {
	find, err := storage.FindByPattern(fmt.Sprintf("*++%s", authD.UserID))
	if err != nil {
		return err
	}

	_, err = storage.Delete(authD.AccessUUID)
	if err != nil {
		return err
	}

	for _, key := range find {
		_, _ = storage.Delete(key)
	}

	return nil
}

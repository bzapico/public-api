/*
 * Copyright (C)  2018 Nalej - All Rights Reserved
 */

package cli

import (
	"context"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-authx-go"
	"github.com/nalej/grpc-login-api-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
)

type Login struct {
	Connection
}

// NewLogin creates a new Login structure.
func NewLogin(address string, port int) * Login{
	return &Login{
		*NewConnection(address, port),
	}
}

// Login into the platform using email and password.
func (l * Login) Login(email string, password string) (*Credentials, derrors.Error){
	c, err := l.GetConnection()
	if err != nil{
		return nil, err
	}
	defer c.Close()
	loginClient := grpc_login_api_go.NewLoginClient(c)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	loginRequest := &grpc_authx_go.LoginWithBasicCredentialsRequest{
		Username:             email,
		Password:             password,
	}
	response, lErr := loginClient.LoginWithBasicCredentials(ctx, loginRequest)
	if lErr != nil{
		return nil, conversions.ToDerror(lErr)
	}
	log.Debug().Str("token", response.Token).Msg("Login success")
	credentials := NewCredentials(DefaultPath, response.Token, response.RefreshToken)
	sErr := credentials.Store()
	if sErr != nil{
		return nil, sErr
	}
	return credentials, nil
}
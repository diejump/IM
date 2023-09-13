package service

import (
	"fmt"
	"github.com/google/uuid"
)

func UUID() string {
	u4 := uuid.New()
	fmt.Println(u4.String())
	return u4.String()
}

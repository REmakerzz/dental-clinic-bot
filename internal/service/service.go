package service

import (
    "os"
    "strconv"
    "strings"
)

var adminUserIDs []int64

func LoadAdminUserIDs() {
    idsStr := os.Getenv("ADMIN_USER_IDS")
    idsSlice := strings.Split(idsStr, ",")
    for _, idStr := range idsSlice {
        idStr = strings.TrimSpace(idStr)
        idInt, err := strconv.ParseInt(idStr, 10, 64)
        if err != nil {
            panic("Invalid ADMIN_USER_IDS")
        }
        adminUserIDs = append(adminUserIDs, idInt)
    }
}

func IsAdmin(userID int64, adminList []int64) bool {
    for _, id := range adminList {
        if userID == id {
            return true
        }
    }
    return false
}
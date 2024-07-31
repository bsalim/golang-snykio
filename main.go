package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// BanList maintains banned IP addresses and their expiration times
type BanList struct {
	m        map[string]time.Time
	mu       sync.Mutex
	duration time.Duration
}

// NewBanList creates a new BanList with a specified ban duration
func NewBanList(duration time.Duration) *BanList {
	return &BanList{
		m:        make(map[string]time.Time),
		duration: duration,
	}
}

// IsBanned checks if an IP is currently banned
func (b *BanList) IsBanned(ip string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	until, banned := b.m[ip]
	if !banned {
		return false
	}
	if time.Now().After(until) {
		delete(b.m, ip)
		return false
	}
	return true
}

// Ban adds an IP to the ban list
func (b *BanList) Ban(ip string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.m[ip] = time.Now().Add(b.duration)
}

// Global instance of BanList with a 10-minute ban duration
var banList = NewBanList(1 * time.Minute)

func main() {
	r := gin.Default()

	// Middleware to check if the IP is banned
	r.Use(func(c *gin.Context) {
		ip := c.ClientIP()
		if banList.IsBanned(ip) {
			c.JSON(http.StatusForbidden, gin.H{"message": "You are banned"})
			c.Abort()
			return
		}
		c.Next()
	})

	r.GET("/banit", func(c *gin.Context) {
		ip := c.ClientIP()
		fmt.Printf("Your IP: %s", ip)
		// Simulate a failed login attempt by banning the IP
		banList.Ban(ip)
		c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})

	r.Run(":8080")
}

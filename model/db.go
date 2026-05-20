package model

import (
	"fmt"
	"net/url"
	"strings"
)

type Mysql struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Other    string `json:"other"`
}

// DefaultOther 默认参数，使用中国时间
const DefaultOther = "charset=utf8mb4&parseTime=True&loc=Asia/Shanghai"

func (m *Mysql) Dsn() string {
	other := m.Other
	if strings.TrimSpace(other) == "" {
		other = DefaultOther
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		m.Username, url.QueryEscape(m.Password), m.Host, m.Port, m.Database, other)
}

type Sqlite3 struct {
	Path     string `json:"path"`
	PassWord string `json:"password"`
}

func (s *Sqlite3) Dsn() string {
	if s.PassWord != "" {
		return fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_synchronous=N&_foreign_keys=ON&password=%s",
			s.Path, s.PassWord)
	}
	return fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_synchronous=N&_foreign_keys=ON",
		s.Path)
}

type PostgreSql struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslmode"`
}

func (p *PostgreSql) Dsn() string {
	sslMode := p.SSLMode
	if strings.TrimSpace(sslMode) == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		p.Host, p.Username, p.Password, p.Database, p.Port, sslMode)
}

type Redis struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

func (r *Redis) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

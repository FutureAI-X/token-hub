package common

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MaxOutboundBodySize 出站响应体最大读取字节数（8MB）。
// 仅靠 http.Client.Timeout 只能限制时长、不能限制体积：
// 恶意或被控的上游可以持续高速输出把进程内存打爆。
const MaxOutboundBodySize = 8 << 20

// outboundDialTimeout 建立出站连接的超时
const outboundDialTimeout = 10 * time.Second

// blockedRanges 额外需要拦截的网段（IsPrivate/IsLoopback 等未覆盖的部分）
var blockedRanges = func() []*net.IPNet {
	cidrs := []string{
		"0.0.0.0/8",       // 本网络
		"100.64.0.0/10",   // 运营商级 NAT (CGNAT)
		"192.0.0.0/24",    // IETF 协议保留
		"192.0.2.0/24",    // TEST-NET-1
		"198.18.0.0/15",   // 基准测试
		"198.51.100.0/24", // TEST-NET-2
		"203.0.113.0/24",  // TEST-NET-3
		"240.0.0.0/4",     // 保留
		"::/128",          // IPv6 未指定
		"64:ff9b::/96",    // IPv4/IPv6 转换
		"2001:db8::/32",   // IPv6 文档用
		"2002::/16",       // 6to4
	}
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		if _, n, err := net.ParseCIDR(c); err == nil {
			nets = append(nets, n)
		}
	}
	return nets
}()

// metadataIPs 云厂商元数据服务地址，SSRF 的经典首要目标
var metadataIPs = []net.IP{
	net.ParseIP("169.254.169.254"), // AWS/GCP/Azure/阿里云
	net.ParseIP("169.254.170.2"),   // AWS ECS
	net.ParseIP("100.100.100.200"), // 阿里云
	net.ParseIP("fd00:ec2::254"),   // AWS IPv6
}

// IsBlockedOutboundIP 判断目标 IP 是否禁止出站访问（内网/回环/保留/元数据地址）
func IsBlockedOutboundIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return true
	}
	for _, meta := range metadataIPs {
		if ip.Equal(meta) {
			return true
		}
	}
	for _, n := range blockedRanges {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ValidateOutboundBaseURL 校验供应商 BaseURL 是否可用于出站请求。
// 该地址会承载解密后的供应商 API Key，因此必须阻断 SSRF：
// 禁止非 http(s) 协议、禁止解析到内网/回环/链路本地/元数据地址。
// 注意：此处校验发生在保存配置时，实际连接时还会有 SafeDialContext 二次校验
// （防 DNS rebinding——保存时解析到公网、调用时解析到内网）。
func ValidateOutboundBaseURL(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return errors.New("地址不能为空")
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("地址格式无效: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("仅支持 http/https 协议，当前为 %q", u.Scheme)
	}
	if u.User != nil {
		return errors.New("地址中不允许包含用户名/密码")
	}

	host := u.Hostname()
	if host == "" {
		return errors.New("地址缺少主机名")
	}

	// 主机名本身是 IP 时可直接判断
	if ip := net.ParseIP(host); ip != nil {
		if IsBlockedOutboundIP(ip) {
			return fmt.Errorf("目标地址 %s 属于内网或保留地址，禁止使用", ip)
		}
		return nil
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("主机名无法解析: %w", err)
	}
	if len(ips) == 0 {
		return errors.New("主机名未解析到任何地址")
	}
	for _, ip := range ips {
		if IsBlockedOutboundIP(ip) {
			return fmt.Errorf("主机 %s 解析到内网或保留地址 %s，禁止使用", host, ip)
		}
	}
	return nil
}

// SafeDialContext 在真正建立 TCP 连接前校验解析出的 IP，
// 阻断 DNS rebinding（配置时解析到公网、请求时解析到内网）。
func SafeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, ipAddr := range ips {
		if IsBlockedOutboundIP(ipAddr.IP) {
			lastErr = fmt.Errorf("拒绝连接内网或保留地址 %s", ipAddr.IP)
			continue
		}
		dialer := &net.Dialer{Timeout: outboundDialTimeout}
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ipAddr.IP.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("目标主机无可用地址")
	}
	return nil, lastErr
}

// NewSafeHTTPClient 构造带 SSRF 防护的出站 HTTP 客户端：
//   - 拨号前校验目标 IP（含 DNS rebinding 防护）
//   - 拒绝跨主机重定向（API Key 会随请求发出，跳转即泄露）
//   - 拒绝重定向降级到非 https
func NewSafeHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:           SafeDialContext,
			TLSHandshakeTimeout:   outboundDialTimeout,
			ResponseHeaderTimeout: timeout,
			DisableKeepAlives:     false,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return errors.New("重定向次数过多")
			}
			origin := via[0].URL
			if req.URL.Host != origin.Host {
				return fmt.Errorf("拒绝跨主机重定向: %s -> %s", origin.Host, req.URL.Host)
			}
			if origin.Scheme == "https" && req.URL.Scheme != "https" {
				return errors.New("拒绝重定向降级到非 HTTPS")
			}
			return nil
		},
	}
}

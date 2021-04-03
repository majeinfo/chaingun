package commands

import (
	"github.com/majeinfo/chaingun/web_proxy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type proxyFlags struct {
	listen_addr string
	proxy_domain string
	proxy_ignore_suffixes string
}

var proxyConfig proxyFlags

var appProxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Starts a HTTP/S Proxy to help you create a Playbook",
	Long: `This mode starts a simple Proxy that intercepts the requests
made by your Web Browser (you must configure it !). At anytime you can
display a simple menu that allows you to stop the Proxy or reset the
captured requests or display the Playbook corresponding to the requests.

Example usage:
player proxy --proxy-domain google.com`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Proxy mode started")
		web_proxy.StartProxy(proxyConfig.listen_addr, proxyConfig.proxy_domain, proxyConfig.proxy_ignore_suffixes)
	},
}

func init() {
	RootCmd.AddCommand(appProxyCmd)
	appProxyCmd.Flags().StringVarP(&proxyConfig.listen_addr, "listen-addr", "", "127.0.0.1:12345",
		"Address and port to listen to (ex: 127.0.0.1:8080)")
	appProxyCmd.Flags().StringVarP(&proxyConfig.proxy_domain, "proxy-domain", "", "",
		"Name of the Domain that must be proxied (ex: github.com)")
	appProxyCmd.Flags().StringVarP(&proxyConfig.proxy_ignore_suffixes, "proxy-ignore-suffixes", "",
		".gif,.png,.jpg,.jpeg,.css,.js,.ico,.ttf,.woff,.pdf", "Comma separated list of request suffixes to ignore")
	appProxyCmd.MarkFlagRequired("proxy-domain")
}

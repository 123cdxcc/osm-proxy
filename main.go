package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"net/url"
	"osm-proxy/cache"
	"osm-proxy/config"
	"regexp"
)

var (
	httpClient *http.Client
	conf       *config.Config
	limiter    *rate.Limiter
	server     = ":"
)

func init() {
	var err error
	conf, err = config.Load()
	if err != nil {
		panic(err)
	}
	if conf.Server.Host != "" {
		server = conf.Server.Host + server
	}
	if conf.Server.Port != "" {
		server = server + conf.Server.Port
	}
	if server == ":" {
		server = ":8080"
	}
	if conf.Cache.Dir != "" {
		err = cache.Init(conf.Cache.Dir)
		if err != nil {
			panic(err)
		}
	} else {
		err = cache.Init("map_cache")
		if err != nil {
			panic(err)
		}
	}
	if conf.Limit.Rate > 0 {
		limiter = rate.NewLimiter(rate.Limit(conf.Limit.Rate), conf.Limit.Rate)
	}
	var proxy func(*http.Request) (*url.URL, error)
	if conf.Proxy.Url != "" {
		proxyUrl, err := url.ParseRequestURI(conf.Proxy.Url)
		if err != nil {
			panic(err)
		}
		proxy = http.ProxyURL(proxyUrl)
	}
	httpClient = &http.Client{
		Transport: &http.Transport{
			Proxy: proxy,
		},
	}
}

func main() {
	err := start()
	if err != nil {
		panic(err)
	}
}

type OSMMapURLBind struct {
	X string `uri:"x"`
	Y string `uri:"y"`
	Z string `uri:"z"`
}

func (o *OSMMapURLBind) Key() string {
	return fmt.Sprintf("%v_%v_%v.png", o.Z, o.X, o.Y)
}

func start() error {
	reg, err := regexp.Compile("/(\\d+)/(\\d+)/(\\d+)\\.")
	if err != nil {
		return err
	}

	r := gin.Default()
	r.GET("/:z/:x/:y.png", func(c *gin.Context) {
		s := reg.FindAllStringSubmatch(c.Request.URL.Path, -1)
		if len(s) == 0 || (len(s) > 0 && len(s[0]) != 4) {
			c.AbortWithError(400, errors.New("param error"))
			return
		}
		var urlParam = &OSMMapURLBind{
			Z: s[0][1],
			X: s[0][2],
			Y: s[0][3],
		}
		data, err := cache.Get(urlParam.Key())
		if err != nil {
			if limiter != nil {
				err = limiter.Wait(context.Background())
				if err != nil {
					c.AbortWithError(500, err)
					return
				}
			}
			//fmt.Println(urlParam.Key(), "缓存无数据，调用远端服务器:", time.Now())
			data, err = download(fmt.Sprintf("https://tile.openstreetmap.org/%v/%v/%v.png", urlParam.Z, urlParam.X, urlParam.Y))
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
			cache.Set(urlParam.Key(), bytes.NewReader(data))
		}
		c.Data(200, "image/png", data)
		return
	})
	err = r.Run(server)
	if err != nil {
		return err
	}
	return nil
}

func download(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "xyz.rsss.map")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

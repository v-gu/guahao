package zjol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/v-gu/guahao/log"
	provider "github.com/v-gu/guahao/provider"
	driver "github.com/v-gu/guahao/provider/driver"
	store "github.com/v-gu/guahao/store"
)

const (
	NAME       = "zjol"
	DOMAIN_URL = "http://guahao.zjol.com.cn/"

	session_file = "session.cache"
)

var (
	domain      *url.URL
	sleepIntval time.Duration = time.Second * 2 // seconds

	stdin  = bufio.NewReader(os.Stdin)
	stderr = bufio.NewWriter(os.Stderr)
)

// A Session tracks a unique login session.
type Site struct {
	User string
	Pass string

	Dept   string // department ID
	Doctor string // doctor ID
	LinkNo int    // booking link index

	hospital string
	patient  string

	session Session
	client  *http.Client
	log.NamedLogger
}

type Session struct {
	SessionId string
	UserId    string
}

type Ticket struct {
	id      string
	no      string
	time    string
	dayPart string

	referer string
}

func init() {
	var err error
	domain, err = url.Parse(DOMAIN_URL)
	if err != nil {
		panic(err)
	}

	provider.Register(NAME, NewDriver())
}

// client should call this function before any other functions or
// methods
func NewDriver() driver.Driver {
	// compose client and session
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	client := &http.Client{Jar: jar}
	site := &Site{client: client, NamedLogger: log.NamedLogger{NAME}}
	return site
}

// Login into account from local stored session information.
func (s *Site) login() (err error) {
	// print patient info
	s.Infof("User: %v, Dept: %v, Doctor: %v, LinkNo: %v\n",
		s.User, s.Dept, s.Doctor, s.LinkNo)

	// read local cache
	err = s.readLocalCache()
	if err != nil ||
		(len(s.session.SessionId) == 0 || len(s.session.UserId) == 0) {
		// local cache contains invaild session info
		return s.realLogin()
	}

	// compose client and session
	jar, err := cookiejar.New(nil)
	if err != nil {
		return
	}
	client := &http.Client{Jar: jar}
	sidCookie := &http.Cookie{
		Domain: domain.Host, Path: "/",
		Name: "ASP.NET_SessionId", Value: s.session.SessionId,
		HttpOnly: true, MaxAge: 0}
	uidCookie := &http.Cookie{
		Domain: domain.Host, Path: "/",
		Name: "UserId", Value: s.session.UserId,
		HttpOnly: true, MaxAge: 0}
	client.Jar.SetCookies(domain, []*http.Cookie{sidCookie, uidCookie})
	s.client = client

	s.Debugf(log.DEBUG_SESSION, "%s: session: %#v\n", NAME, &s.session)
	return
}

func (s *Site) readLocalCache() (err error) {
	err = store.Store.Unmarshal(&s.session, NAME, session_file)
	s.Debugf(log.DEBUG_CONFIG, "problem reading session information: %s\n", err)
	return
}

// Account login function, a real connecting login
func (s *Site) realLogin() (err error) {
	// get login VfCode
	vfcode, err := s.getVfcode("http://guahao.zjol.com.cn/VerifyCodeCH.aspx",
		"storage/zjol/logincode.jpg")
	if err != nil {
		return err
	}
	for _, cookie := range s.client.Jar.Cookies(domain) {
		if cookie.Name == "ASP.NET_SessionId" {
			s.session.SessionId = cookie.Value
		}
	}

	// login
	loginUrl := fmt.Sprintf(
		"http://guahao.zjol.com.cn/ashx/LoginDefault.ashx?idcode=%v&pwd=%v&txtVerifyCode=%v",
		s.User, s.Pass, vfcode)
	s.Infoln("logging in...")
	s.Debugf(log.DEBUG_HTTP, "GET -> '%v'\n", loginUrl)
	resp, err := s.client.Get(loginUrl)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	s.Debugf(log.DEBUG_HTTP, "responce -> '%s'\n", b)
	loginResp := strings.Split(string(b), "|")
	if len(loginResp) <= 1 {
		return errors.New(loginResp[0])
	}
	s.session.UserId = loginResp[1]

	// store session information to cache
	err = store.Store.Marshal(&s.session, NAME, session_file)
	if err != nil {
		s.Warningf("problem writing session information to cache file: %s\n", err)
	}

	// compose cookies
	uidCookie := &http.Cookie{
		Domain: "guahao.zjol.com.cn", Path: "/",
		Name: "UserId", Value: s.session.UserId,
		HttpOnly: true, MaxAge: 0}
	s.client.Jar.SetCookies(domain, []*http.Cookie{uidCookie})

	s.Debugf(log.DEBUG_SESSION, "session: %#v\n", &s.session)
	return
}

// book a ticket with n's booking link in the page. 'n' starts from
// '1'. May panic if session is expired.
func (s *Site) Book() error {
	// 科室
	s.Infof("entering department...")
	link, err := s.getDivUrl()
	if err != nil {
		return err
	}

	// booking loop
	for {
		if end, err := s.loop1(link, s.LinkNo); end {
			return err
		}
		if err != nil {
			s.Errorln(err)
		}
		stderr.WriteString("failed to accquire booking link, press ENTER to retry:")
		stderr.Flush()
		stdin.ReadString('\n')
	}
}

func (s *Site) loop1(link string, n int) (end bool, err error) {
	s.Debugf(log.DEBUG_HTTP, "GET: '%v'\n", link)
	resp, err := s.client.Get(link)
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	sb := string(b)

	// check response body to see if session was expired
	if sb == "-2" {
		s.Warning("session expired")
		for {
			if err = s.login(); err != nil {
			}
		}
	}

	// grub all bookable links
	re := regexp.MustCompile(`showDiv\('(.*)'\)`)
	// TODO: should match exactly, e.g. index -2
	sigMatches := re.FindAllStringSubmatch(sb, -1)
	if sigMatches == nil {
		// can't find match, retry
		return false, errors.New("can't find matching booking link")
	}

	// select n's link
	if n > len(sigMatches) {
		return false, errors.New(
			fmt.Sprintf("link number[%d] > maximum[%d] availible link choices",
				n, len(sigMatches)))
	} else if n == 0 {
		return false, errors.New(
			fmt.Sprintf("link number[%v] should start from '1'", n))
	}
	sig := sigMatches[n-1][1]

	// list 号源
	fd := url.Values{"sg": {sig}}.Encode()
	s.Debugf(log.DEBUG_HTTP, "POST data -> '%v'\n", fd)
	req, err := http.NewRequest("POST", "http://guahao.zjol.com.cn/ashx/gethy.ashx",
		strings.NewReader(fd))
	req.Header.Add("Referer", link)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	s.Debugf(log.DEBUG_HTTP, "header -> %V\n", req.Header)
	resp, err = s.client.Do(req)
	if err != nil {
		return false, err
	}
	b, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return false, err
	}
	s.Debugf(log.DEBUG_SESSION, "showdiv('%v') -> %s\n", sig, b)

	// numbering plan
	lists := strings.Split(string(b), "#")
	hospital := lists[1]
	dept := lists[2]
	doctor := lists[3]
	patient := lists[6]
	numbering := lists[11]
	sig2 := lists[12]
	s.Debugf(log.DEBUG_SESSION, "%v %v %v %v %v %v\n",
		hospital, dept, doctor, patient, numbering, sig2)
	nums := strings.Split(numbering, "$")
	nums = nums[1:]
	fmt.Printf("[%v] %v %v %v:\n", patient, hospital, dept, doctor)
	tickets := make([]*Ticket, len(nums))
	for i, num := range nums {
		fields := strings.Split(num, "|")
		t := &Ticket{
			id:      fields[0],
			no:      fields[1],
			time:    fields[2],
			dayPart: fields[3],
			referer: link}
		tickets[i] = t
		fmt.Fprintf(stderr, "\tNo:%v Time:%v DayPart:%v Code:%v\n",
			fields[1], fields[2], fields[3], fields[0])
		stderr.Flush()
	}

	// call booking loop
	for {
		end, err := s.loop2(sig2, tickets)
		if end {
			return end, err
		}
		if err != nil {
			s.Errorln(err)
		}
	}
}

func (s *Site) loop2(sig2 string, tickets []*Ticket) (end bool, err error) {
	var ticket *Ticket
	var n int
	if len(tickets) > 5 {
		n = len(tickets)/2 + 1
	} else if len(tickets) > 3 {
		n = 3
	} else if len(tickets) == 3 {
		n = 2
	} else if len(tickets) == 2 {
		n = 1
	} else if len(tickets) == 1 {
		n = 0
	} else {
		return true, errors.New("all numbers booked!!")
	}
	ticket = tickets[n]

	// get booking vfcode
	key := int(rand.Float32() * 10000)
	u := fmt.Sprintf(
		"http://guahao.zjol.com.cn/ashx/getyzm.aspx?k=%v&t=yy&hyid=%v",
		key, ticket.id)
	s.Debugf(log.DEBUG_SESSION, "ticketing url -> '%v'\n", u)
	resp, err := s.client.Get(u)
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	// TODO: remove magic string
	ioutil.WriteFile("storage/zjol/bookingcode.jpg", b, os.ModePerm)
	stderr.WriteString("ticketing vfcode:")
	stderr.Flush()
	vfcode, err := stdin.ReadString('\n')
	if err == nil || err == io.EOF {
		// re-get vgcode if a empty line was read in
		if len(vfcode) == 1 {
			return false, errors.New("re-download vfcode...")
		}
		vfcode = vfcode[:len(vfcode)-1]
	} else {
		return
	}
	s.Debugf(log.DEBUG_SESSION, "ticketing vfcode input: '%v'\n", vfcode)

	// book a ticket
	fd := url.Values{
		"lgcfas": {ticket.id},
		"yzm":    {vfcode},
		"xh":     {ticket.no},
		"qhsj":   {ticket.time},
		"sg":     {sig2}}.Encode()
	s.Debugf(log.DEBUG_SESSION, "ticketing POST data -> '%v'\n", fd)
	req, err := http.NewRequest("POST", "http://guahao.zjol.com.cn/ashx/TreadYuyue.ashx",
		strings.NewReader(fd))
	req.Header.Add("Referer", ticket.referer)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	b, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	bs := string(b)
	s.Debugf(log.DEBUG_SESSION, "ticketing response -> '%v'\n", bs)
	if strings.HasPrefix(bs, "ERROR") {
		tickets = append(tickets[:n], tickets[n+1:]...)
		err = errors.New(bs)
		return
	}
	s.Infof("success -> '%v'\n", bs)

	return true, nil
}

// return a URL string provided booking address
func (s *Site) getDivUrl() (url string, err error) {
	if s.Dept == "" {
		return "", errors.New("no valid department info")
	}
	if s.Doctor == "" {
		return fmt.Sprintf("http://guahao.zjol.com.cn/DepartMent.Aspx?ID=%v", s.Dept), nil
	} else {
		return fmt.Sprintf("http://guahao.zjol.com.cn/DoctorInfo.Aspx?DEPART=%v&ID=%v",
			s.Dept, s.Doctor), nil
	}
}

func (s *Site) getVfcode(url string, outputPath string) (vfcode string, err error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	// TODO: remove magic string
	ioutil.WriteFile(outputPath, b, os.ModePerm)
	stderr.WriteString("vfcode:")
	stderr.Flush()
	vfcode, err = stdin.ReadString('\n')
	if err == nil || err == io.EOF {
		vfcode = vfcode[:len(vfcode)-1]
		s.Debugf(log.DEBUG_SESSION, "vfcode input -> '%v'\n", vfcode)
	}
	return
}

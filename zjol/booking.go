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

	glog "github.com/golang/glog"
)

const (
	domainUrl = "http://guahao.zjol.com.cn/"
)

var (
	domain      *url.URL
	sleepIntval time.Duration = time.Second * 2 // seconds

	stdin  = bufio.NewReader(os.Stdin)
	stderr = bufio.NewWriter(os.Stderr)
)

// A Session tracks a unique login session.
type Session struct {
	SessionId string
	UserId    string

	Dept   string // department ID
	Doctor string // doctor ID

	hospital string
	patient  string

	client *http.Client
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
	domain, err = url.Parse(domainUrl)
	if err != nil {
		glog.Fatalln(err)
	}
}

// Login into account from local stored session information.
func Login(sessionId, userId string) (*Session, error) {
	if sessionId == "" {
		return nil, errors.New("session_id is null")
	} else if userId == "" {
		return nil, errors.New("user_id is null")
	}

	// compose client and session
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}
	sidCookie := &http.Cookie{
		Domain: domain.Host, Path: "/",
		Name: "ASP.NET_SessionId", Value: sessionId,
		HttpOnly: true, MaxAge: 0}
	uidCookie := &http.Cookie{
		Domain: domain.Host, Path: "/",
		Name: "UserId", Value: userId,
		HttpOnly: true, MaxAge: 0}
	client.Jar.SetCookies(domain, []*http.Cookie{sidCookie, uidCookie})

	session := &Session{
		SessionId: sessionId,
		UserId:    userId,
		client:    client}

	return session, nil
}

// Account login function, call this if no session informations could be restored.
func RealLogin(user, pass string) (*Session, error) {
	// compose client and session
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}
	session := &Session{client: client}

	// get login VfCode
	vfcode, err := session.getVfcode("http://guahao.zjol.com.cn/VerifyCodeCH.aspx",
		"storage/zjol/logincode.jpg")
	if err != nil {
		return nil, err
	}
	for _, cookie := range session.client.Jar.Cookies(domain) {
		if cookie.Name == "ASP.NET_SessionId" {
			session.SessionId = cookie.Value
		}
	}

	// login
	loginUrl := fmt.Sprintf(
		"http://guahao.zjol.com.cn/ashx/LoginDefault.ashx?idcode=%s&pwd=%s&txtVerifyCode=%s",
		user, pass, vfcode)
	glog.Infof("logging in...\n")
	glog.V(2).Infof("GET: '%s'\n", loginUrl)
	resp, err := client.Get(loginUrl)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	glog.V(3).Infof("response -> '%s'\n", b)
	loginResp := strings.Split(string(b), "|")
	if len(loginResp) <= 1 {
		return nil, errors.New(loginResp[0])
	}
	session.UserId = loginResp[1]
	uidCookie := &http.Cookie{
		Domain: "guahao.zjol.com.cn", Path: "/",
		Name: "UserId", Value: session.UserId,
		HttpOnly: true, MaxAge: 0}
	client.Jar.SetCookies(domain, []*http.Cookie{uidCookie})

	return session, nil
}

// book a ticket with n's booking link in the page. 'n' counts from 1.
func (s *Session) Book(n int) (err error) {
	// 科室
	glog.Infoln("entering department...")
	link, err := s.getDivUrl()
	if err != nil {
		return err
	}

	// booking loop
	for {
		err = s.loop1(link, n)
		if err != nil {
			glog.Errorln(err)
		}
		stderr.WriteString("booking link not availiable, press ENTER to retry:")
		stderr.Flush()
		stdin.ReadString('\n')
	}
	return
}

func (s *Session) loop1(link string, n int) (err error) {
	glog.V(2).Infof("GET: '%s'\n", link)
	resp, err := s.client.Get(link)
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	sb := string(b)

	// grub all bookable links
	re := regexp.MustCompile(`showDiv\('(.*)'\)`)
	// TODO: should match exactly, e.g. index -2
	sigMatches := re.FindAllStringSubmatch(sb, -1)
	if sigMatches == nil {
		// can't find match, retry
		return errors.New("can't find matching booking link")
	}

	// select n's link
	if n > len(sigMatches) {
		return errors.New(
			fmt.Sprintf("link number[%s] > maximum[%s] availible link choices",
				n, len(sigMatches)))
	} else if n == 0 {
		return errors.New(
			fmt.Sprintf("link number[%s] should start from '1'", n))
	}
	sig := sigMatches[n-1][1]

	// list 号源
	fd := url.Values{"sg": {sig}}.Encode()
	glog.V(2).Infof("POST data -> '%s'\n", fd)
	req, err := http.NewRequest("POST", "http://guahao.zjol.com.cn/ashx/gethy.ashx",
		strings.NewReader(fd))
	req.Header.Add("Referer", link)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err = s.client.Do(req)
	if err != nil {
		return err
	}
	b, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	glog.V(3).Infof("showdiv('%s') -> %s\n", sig, b)

	// numbering plan
	lists := strings.Split(string(b), "#")
	hospital := lists[1]
	dept := lists[2]
	doctor := lists[3]
	patient := lists[6]
	numbering := lists[11]
	sig2 := lists[12]
	glog.V(2).Infof("%s %s %s %s %s %s\n", hospital, dept, doctor, patient, numbering, sig2)
	nums := strings.Split(numbering, "$")
	nums = nums[1:]
	fmt.Printf("[%s] %s %s %s:\n", patient, hospital, dept, doctor)
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
		fmt.Fprintf(stderr, "\tNo:%s Time:%s DayPart:%s Code:%s\n",
			fields[1], fields[2], fields[3], fields[0])
		stderr.Flush()
	}

	// call booking loop
	err = nil
	end := false
	for {
		end, err = s.loop2(sig2, tickets)
		if end {
			break
		}
		if err != nil {
			glog.Errorln(err)
		}
	}
	return
}

func (s *Session) loop2(sig2 string, tickets []*Ticket) (end bool, err error) {
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
	glog.V(2).Infof("ticketing url -> '%v'\n", u)
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
		vfcode = vfcode[:len(vfcode)-1]
	} else {
		return
	}
	glog.V(2).Infof("ticketing vfcode input: '%s'\n", vfcode)

	// book a ticket
	fd := url.Values{
		"lgcfas": {ticket.id},
		"yzm":    {vfcode},
		"xh":     {ticket.no},
		"qhsj":   {ticket.time},
		"sg":     {sig2}}.Encode()
	glog.V(2).Infof("ticketing POST data -> '%s'\n", fd)
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
	glog.V(3).Infof("ticketing response -> '%s'\n", bs)
	if strings.HasPrefix(bs, "ERROR") {
		tickets = append(tickets[:n], tickets[n+1:]...)
		err = errors.New(bs)
		return
	}
	glog.Infof("success -> '%s'\n", bs)

	return true, nil
}

// return a URL string provided booking address
func (s *Session) getDivUrl() (string, error) {
	if s.Dept == "" {
		return "", errors.New("no valid department info")
	}
	if s.Doctor == "" {
		return fmt.Sprintf("http://guahao.zjol.com.cn/DepartMent.Aspx?ID=%s", s.Dept), nil
	} else {
		return fmt.Sprintf("http://guahao.zjol.com.cn/DoctorInfo.Aspx?DEPART=%s&ID=%s",
			s.Dept, s.Doctor), nil
	}
}

func (s *Session) getVfcode(url string, outputPath string) (vfcode string, err error) {
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
		glog.V(2).Infof("vfcode input -> '%s'\n", vfcode)
	}
	return
}

func debugHeaders(header http.Header) {
	glog.V(3).Infof("======== cookies =========\n")
	for k, v := range header {
		glog.V(3).Infof("%s -> %s\n", k, v)
	}
	glog.V(3).Infof("======== cookies =========\n\n")
}

func debugCookies(cookies ...*http.Cookie) {
	glog.V(3).Infof("======== cookies =========\n")
	for _, cookie := range cookies {
		glog.V(3).Infof("%s -> %s\n", cookie.Name, cookie.Value)
	}
	glog.V(3).Infof("======== cookies =========\n\n")
}

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDir struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	data []byte
	once sync.Once
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDir) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDir{fs: _escLocal, name: name}
	}
	return _escDir{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadAll(f)
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/css/dns.css": {
		local: "static/css/dns.css", size: 559, modtime: 1379711203,
		compressed: `
H4sIAAAJbogA/1SRTY7bMAyF1/EpWATdBFF+0KILB+gBeolCtmhbGFnUSBQcI8jdh7I9mMzKML/H90Ty
NNh+UO8Z46yiZoQHdORZTSh1rqEhZ+BZnZKjSUVMgXwqopYcxRoiFlqdD/vEmnP6b/IYBBubgtNzDZ48
wvNwrqqBR3cUPzMLHzb76+Xy81YcWDeu+C5fJa2UBXf2jmbhHRFjhEe1C9oY6/saLtId7tvnVu0aigaj
YgriKyCRswb2bdsK3N47DZaxaHX71kfK3qiN7BEFfCX9hVDSRh1761VDzDSK73WJ2qpr1K9SWpZwmqIO
YX3mKPzblNXu81dnJvhhx0CRtecXsglX+zLhIlW/14iyxfPhX04ss40IOjaWo44zpKBbBJHKbqBBnhA9
bINob+RM0kMd8IByOc/o+Si7EJFUZjGSsiM5rETIGCGn4eVIf0o8PEHoRwAAAP//1THiwC8CAAA=
`,
	},

	"/js/colors.js": {
		local: "static/js/colors.js", size: 3381, modtime: 1379711203,
		compressed: `
H4sIAAAJbogA/7xWzW7bOBC+5ykGBRaQU8uS1SYo4rjdbA+bBXYv2dyCoKAkWmIiUQpJuS2KvPvOUKJt
MUr6c1gBicX5+TjDmfmo6Pj4CI7hiptOSQ2m5KBbnomN4DnIrk65gmYDWVM1SgPfcll9RQvFWQ6NtPZW
Z72M6mpEI8DfW6ZYDQB/SYP/JdBzjdZPMHOuheK5dVI2DIALpdhX6J8LCcwus0YaJqSQxX5bDUKCKlLY
NKpeIEZ0tGUKCi65YoZ/7G3WEGw6mRmBEQcz+HZEuK86jbkaJTLzanVkRVEEpTHtWRTVd+KOZfe6kYus
qaMkjt9FcRLhTqFpwlJXIZN5uFtuQxtNWDc5r/BdbrnSuFvIqqJRwpS1DoUM79iW6UyJ1vTb0dnTcwwf
rYvRwCRc/fnHcKZbVnUcTAOX//69GGwI1mbbVcx5s5y1Buu1UU3tMuBy8Vnci5bngi0aVUS0ihDok8X+
pFuW8YVDuNC6q7kGNYdijkHkkOKpc3fmiC36Ymtu4CaeQ3Jyckt2DkANDVTOQfcA1ZSL29C57drEVZxq
6VoFuwLDUZzLOaR0Egenovum2OFMdY7DwaQRC3tWc8yGumBwi+wvNQxue91c6gpbZd8pKHTNsrNDAxTf
xLcQUULzo4O9oBi0y0ltOmiTSW3NvqD+H2bKBb4GfSHSmWeER+qMhHzGKEcTQgvJfKwqx0s9XlL2AXm+
Js8ZRTkMBj1i0yvX63Wv/jaGRmeatHhFY8SyEnsRzzrbGT0Cr3Dixl7kUcF7iBcn8AEjjyBIKHAX/gzO
euk+rNUY4LMwWWn1fkQZw+3U2UjmIg0KxE8pxRxhcXWO5fkAp7hd7O1AT4p8d796il48g54iunLoyY/j
pc/gKcQrHN7bH8J79KoTreF0b/G4r+swOTe2q1TTyTwocaKwQWdzOBDqKWE1CG976EdHpNPMRoPoMRuS
3f/DbIfE9EvMdkiNv8JseGlYFTFS2XEMhyGwpaM5VKIojeRa/zy70XXxHXbDra+bqyIdsRsKfXajXkPx
M+ymB+00u1WDdprdHuycnw9zTl0TLLGVNc13RS/Y4CT1GKlFtwTFFaofxirlUa/HtR7vdTxRXv7tHCHB
+KxBD1GdoWhRCQZer2G58gZqb/Yelr1Z+KLZOSzxYE7JdKhmS8zzgJm1M0zxFP/M99yTA/eHl4wTNH7z
7F6BVePK2J2ngJyfRyrefaDtbRD7Z0i3JN2FdODVT90H5DkUayhQiZEvbTarJ3etZ+lZpBNY4ROsl5lQ
TZFeMSVMp5lwwNy3nfSHLkXit9+p/eeB/QwJbpZv3s2B/pLk9Na74MnjEulzvffFkfWG1XAaniB5Gy9i
TFl6GGK8lPyLsUBjcea+n29uDyrfSxdtp8vAUUuwC2U2OzBFOodAgJ0MgX0p6YfmyW+ZXQCHWa2mTYif
6JIdjuF1n+wxiBn8Bjbh1UQWXrw7tNmLzdD77mr6GJD5fwEAAP//ZXB4JDUNAAA=
`,
	},

	"/js/dns.js": {
		local: "static/js/dns.js", size: 2514, modtime: 1441656343,
		compressed: `
H4sIAAAJbogA/5xWT2/bNhS/+1M8cEFFdY4cD9klbjJg22U7rBt2LAqDkWiJLSWyJJXMM/zd9yiRsqQo
WFFeTJPv7++996NIazlYZ0TuyG61ooe2yZ1QDdCrFE6rFeB6Ygby1hjeuL1WWj1xs7vctLpgjsM9RFXq
FSGsq6zk7ve/3/9Bkw3TYmMdc61N1hfp/mSs49dmk6vGKskzqUpKcrKGILibyPkILDcYksUQepEsHOxm
FmGfHYR03NCp3DgYjAMMd61pAO97sZq5vKJkm93ekBTO6WpitzRMV5hlww3i8IuSylj6/vETz132mR8t
DV7STPKmdFW6m+pf0eS7mIHVrPlguLx3Skkn9MckzcKWJpUoeDJLf6zsHlVxRIXK1ZISMpPcZ5xhFvhr
lXE/H+ly8hMdv8QBqM0aVmOJ7+8B7V4QIv/2i8D5hd4Aoled3J7T//PZQ2p4rkwRnGP5sy96Xn2/+ntf
/T1WSgf5cV7NKOQT+Ns7aOCMgSxZy30F0Vqsq+tKGswuamBc+1wy23Wg/wdv3oTNA2x/vIGfgFSirK6/
tNwcr32bELhDKF8xtq2joW2N5f8TgRDWJ3K76N5wq3FW+N6Jmg+B0NlFH9P06AFub25SH56V6vk6Xr4W
nJ81x2st+3mP2zhHaLwpcLhOYSDvwC5CPOpZ7FamNarRaGwmf55PC07xpDuIbeuamePAD1k4WOgW1O18
9wJxUsZ5BNWQSLAIKXzrzJ6QWMvSQ0EqT5sYJbrKeY1Uimd4WTnyIseO05zvQU+cmWfnphSHY6CtNbQY
3kE0vFjDD0uE0Inti7bWMUm0MRKMVTkHxwGYgj+25d6pspRINBk2XJJLkX8eszWP84o6y/56fRpcoNiQ
37TwY/OEZUKTJTd+8Uwb/oSQ/coPrJWOjlLxWAmNUF1RVwkkWcf/mQhMn5HpK4YeZyevKvpAeQH4MHaB
Cr2Gzh9cFDxTzs0h38yOsvAb+XxEnTGFKHGCS/vUrGmZRM/ECSc5waODUmQyYHN9Yiv1PH4HZqFcUAsF
Giq1eTsvWNfmWKRkFN4YKtd19OInwFeVya9A0ck7LPdD8oI3vl8eh/iKf8DRm43Ggonk3aazPhqGIXX/
+3YzjEREqeC1CnAQ1k16uPlIoocBk+Ggaxj6tf0crvynRbftv6kiPs+iKdQz5ul+Q5zNE5O0F1jDdosE
vludU/rpL/+44P6/AAAA//8afmmt0gkAAA==
`,
	},

	"/js/graph.js": {
		local: "static/js/graph.js", size: 1676, modtime: 1379711203,
		compressed: `
H4sIAAAJbogA/6RUPW/bMBDd9SsOWiS5guUWyBJDS9MOGdoOzhZ4YKSzxFYiVZKKayT+7yUp2vqyCwPl
INi8u3d3794xWSw8WMA3wkiBElSJsKk5VyXFh5IIJaGiL4KIA+ypKq1d7TkUgjSlhD0CEQgFMhREUVYA
Z9aHshz/QKMxDXqpVHOfJNIBZxZ4yUWhjYn3SkSHBymEu5ZlimqUMII3D/TxW4kglaCZ8teevTIRJ7AN
ilcUUscy3I9rD6MYkgR+MNRN6E+DOsy6gy3BYpmjaG1wqCbAnhTejvHZesr0xBWprPVapk1b14aqa+gd
Qhf/dL7VwX0uW94Dr7iQ80qy8/2pzueto2RCx1IThqR+4mHOs7ZGppYFqq8Vmp+fD4956FvKnbuv6/+4
Wq0gWeRYkcMiidberPdbQa3zjZAkzwdMTKiKXHMCVSuY04PVhMCMi9y/h14vHXXfSY0x/G5kNHA3h+5g
AP/ce28hTVNotWB3WiT5NG48wVEcXBjlehZsxGrkZ4c6NKRunktZ0p2WELy/w/Onu7sYTp/tHG0okHET
fZILURN1XGN9iBhfIMJiKcF/4UYdKryHQBQvYQAf+uTLn5yyMIiDSN8GUTDDOE44Os7G1BdhR+PLbq38
f4/G6alpkOWhGcsXojCMjERNr2ZFGyIkPjIVan0YdUbTUgAr/djcKoD/THb+N9hw3z2lHZdyJHB2SdJO
QhWyQj/Phq7VJZrcy5HCGF9j3laUsv5X9m2a0K3rNan2Gbtsx7V3DA07fwMAAP//EDGmY4wGAAA=
`,
	},

	"/js/templates.js": {
		local: "static/js/templates.js", size: 2318, modtime: 1441705446,
		compressed: `
H4sIAAAJbogA/7RWTXObMBA9t79io0MHpgoEYyeOv0499Af0Vnc8MsiOpiAUaUkn4/i/V4BNYxk3dj8y
w0TKwnv7dqW3ESvwrq6ukOcqY8iND09MQ7uFKWy24/ft/isxXD9xTb7ZiOQ/4HOxZjL4sot7m6RI+QhW
pUxQFBK8hCoqfNg0sFN8EGaMwdITU/HyQohfb8gE9Wy/nksCH0H4Y7HyMDD2WXl7VlrBRX7964ZGt3QQ
DynZbGC7Jb6/wUCbmnHPX2+wCtQk6WxiFJNg8Dnj0zlZsuT7WhelTK+TIiv0SK/te00eGDw11HWkYb7x
/V2S/pzMPsilUeNJWEHOoEHWPLO4WBQZCjUngAJrJnIoR7KcG1dNNKRR7+5cOW1+FZabHli+rT9OAlUo
zy6bwpKZq63r252gSWjL9aonbnfakjrSPjFkrrK7Po3uo4uVJYWUvA4vRPoqzSNlZ+faNCnJmDG2KVWH
jo9d8zKDB81X9qUHRDUKQ7dyQrl1Gw1vBnFokGFpOkp9/MEkZKfYz9TTKnHJHpVZ1DGX86RgsD9OJy2I
08i416NxHF/cyBbpVyJh9beuQ/pHhTgh6t2xoih3Nd0OaHx3tou81rTH+t+qWk5rgcbmc3yQTlagOfQ5
yzL3nla+d9TgfnxP+/3BZcVIPRKcZUDWWepM3sjXVV0qFDlfdNyfi2DsdcBFqVI7pv4Sqbnjb4G4cF2u
VU89zbHUEix0ZmOwpYppFCwzIzt7KZhyWa3ANsQ/nMRlnjP9/O9G8cHA3YE7Tt6jw/MHbuUph467RAn2
uRZyVdSLjOk173DLDseAx5JrYf8lUVyD4XY4pLtR9dtSX1TdnwEAAP//pWZFrg4JAAA=
`,
	},

	"/js/templates2.js": {
		local: "static/js/templates2.js", size: 2146, modtime: 1441705010,
		compressed: `
H4sIAAAJbogA/7RV3XKbPBC9/r6n2Oiig6YKxHab+P+6D9C7upORQTiaglCkJZ0M8btXgE1jGTd2fzyj
scSKc/bssrsyheDq6gpFrjOOwlJ44ga6Iyyg2s7+785fiBXmSRjy1VmU+A6fig1X4eedPajiIhFTSEsV
oywUBDHTTFKoWtgFPkg7w3AdyIV8eSGENgcyR7Pc71eKwHuQdCbTAEPrVhrsWVkNN6DN3w0b3LIPkyEj
VQXbLaG0wtDYhnHP3xywNjQkyXJuNVdg8TkTixVZ8/jbxhSlSq7jIivM1GzcvdYPDJ9a6sbSMt9QunOS
rsjynVpbPZtHNeQSWmQjMoeLRZGh1CsCKLFhIodyFM+F9dUMxmwwvDtXTudfjeW7B45vS2dxqAsduG0b
WLL0tfW9uxM0j1y4XuXEz44X0jjj1jqhterjVLaXOTwYkbpLD4h6GkW+N1L7vkzHNx9HkUWOpe1x//iF
ecRPsZ+pp1Pikz1qe9/YfM6TgsH9vMQ7EC/tw7tbNhxfnvYO6acjUf2sL/G/FYgTov47VjTIPU2jwYSN
RmdX5mtNe6x/rarjdG3FOn+OP6STEWg/+pxn2dKLRd1LjhI8mozcmlwWjCQg4VlF7aq18eQNf33VpUaZ
i/ue+rkIxpUD3pc6ca3/D5HaGn8LxIfrCUc7SYzA0ihw0JmzwZZpblDyzE7dPGNgy3W9A5cQejjdyjzn
5vnvjbeDIbYD9/r+kI3HZw+xuqccdtw1KnDrWqq0aDYZNxvR0y17OgY8lsJIN+a1MGBFXKhk1/5/GeqL
ovsjAAD//99994BiCAAA
`,
	},

	"/templates/client/server.html": {
		local: "templates/client/server.html", size: 556, modtime: 1441705443,
		compressed: `
H4sIAAAJbogA/1RSTW/kIAw9b35FxEqr7aGlVVWpmhJO/R8Rk9CZqAQoOJUqi/9eO0zmI4dgHs/PzwYF
STeIf7NN3zaV0igYtcrR+DbDj7Od2Jvh85DC4sf7IbiQdumw/4+4xqXcCf3P73N8U5KTdFtzk3WdgBAc
TFG0MAErURlvZptLQeSglBZRniChN/CkpCQ5aVY/lPhuwHDeELy3A0zB99PIgKwnV+zqYHAm505Qed0o
0x6T/ejEESDupEScYim718eXZ5nBwJK5OmNKGqKftTYVxK+Y+3XDTpuWPjJFIFtYF0l/crPGVwq6+VOJ
T/OJysGFzLvmplW6hkztbS3VjmbjHI+BL6IWfajT2wAa2sq5JCEuEabZ9vFaCpGagH6JowF7e1AncR4l
iV8eheRn8hsAAP///MEK9CwCAAA=
`,
	},

	"/templates/client/summary.html": {
		local: "templates/client/summary.html", size: 101, modtime: 1379711203,
		compressed: `
H4sIAAAJbogA/6quVi4uzc1NLKqsreVSAAKb4oLEPIXknMTiYlulpJI8BSDWzcxLywczchKL0lOV7Kqr
CwuKa2sVCktTizJTixUKUosUilOT8/NSbPRB+u24qqv1EeYCAgAA//+eRmtkZQAAAA==
`,
	},

	"/templates/index.html": {
		local: "templates/index.html", size: 3692, modtime: 1441705316,
		compressed: `
H4sIAAAJbogA/6RXbW/bthN/309x1f9dUUlx+pA2lf3m36It0GHd2g0YimKgxbPEhCJZknLqDdtn35GU
LaV2EwcNYIcPv/vd8Xh3PFf3X/78/49/vH8Fre/k4l4V/oFkqplnqLLFPYCqRcbDgIYdegZ1y6xDP896
v8qfZcOWF17iokHNlYNOK+G1rcq0OhFWrMN5thZ4ZbT1GdRaeVREdiW4b+cc16LGPE4egiAWwWTuaiZx
PitOsn0qjq62wnih1YTtAJD1vtX2OiaB7uc5vENwfiPRQZ4PslKoS2gtruZZ6705L0vnCyO6plDoy5qr
UoqlK5dae+ctM+VpWbvJvOiEKmglA4tyniX6FtFvjYsraQyw1HwDfw8TAMM4F6rJvTbn8PTEfH0B5YM4
AK+hY5cIvsV4GCYUWmg0MCnj4hXbBFAYLrX3ugO9ijMiWzILD8pBzT/JjnJiyA+fOrfojFZOrPEmB+wp
IzXMizqSUQjdLDfc2JuPP717Aq4V3UNYaQtvXz3Nn4HrTYitcOgEQIkd3fjkakn+k1iB9CQCzz9v76BK
oQTO1mRQGXLhSWAvGq0bibXmWNS6K91ald726jJBiguXLciLUXir4RMqLlafg8prNq/YGpjidBd93YKg
C/w25NKhKVZ93fuIyAYfiY416EqiCKsFfWV7gswYiXlkzw/IfrtdGNUcQ+LEX+jm2dnp17PTWyjzCLoj
8Wz2+Ct9bqMeYDvyqkzVKQxDBg0KuVhDLZkj4iuKSYN2FzphS/B51lHaZFvQLo0CrOrldl3RZdEn92zp
kj4pFhUbjPxfqzvMgDPPyMaGImSeETJbvKH1qmQUEwTfk2rIovaQ2Ouw8X05t3EeO3dI8kPa+r4sx2Xf
HJJ8GTZGuars6RUgH0w8SLh8KJrZvn/DrmEKgdWecj6Lzo2OGTw+8bnru47ZTUgWWhr3iUPihJAm8Tuo
5agc8mFOFUYY5EmJQ7tG67Jd9vrxpQozu9hV0zDlqcAPL805nFEtJc9FEnqo+M3opxH99uXtyNlpgr6/
HfosIn/p0Qp0x8L/nVFZhS/mCIF0xN/JSfQ6Hsv/2ur+GPKE/hWpbluP/FhrfjMUhMfAZycRPwXS2E7G
1687pf+4O5aDOA3xs60Bk+g7FMopvFKajjZWZmvgipIhD3Xr3HX05r6AtBA2z4VnUtRDLjtgFoGjZBsK
4eUGNOWJwxDVRVWaCXfN1Jq5Ue+HIbYhtUXZ85MTKo0ompYal8c0Ib8kmRtIPmoy5jDFbJ9i65Vb3LKt
QqPeofhQt0ZvcMdCJ1bcwc+pNO1uypDHoh5qB3r3J+87M+p6p1noioqi2F0s4Q/ZHyfp0S3HJim+xtff
4wtGHogvd1ye3/nvWlcB7yWr6a6Zjw0X9QDb3ovrug9dCLjUmZnwwIGkA1FLQA60I8+0C7mpB7v4QpVj
U86Ks2I2TGLLdaAdOZZy2tZdfNvL/gBvT3XculpbJHMfkbnjQr5PHT15J7NZfbmk5CLy50S+ne5T393J
rW6YKh8VJ8VpGh/jiqGXJQ9SZhhJBc/9iPdcR/fQCjrebnjAazdYUWup7a0mjPhYPo6Hh379Oji2ZqkC
V2X6aXnvvwAAAP//MdPKeWwOAAA=
`,
	},

	"/": {
		isDir: true,
		local: "static",
	},

	"/css": {
		isDir: true,
		local: "static/css",
	},

	"/js": {
		isDir: true,
		local: "static/js",
	},

	"/templates": {
		isDir: true,
		local: "static/templates",
	},

	"/templates/client": {
		isDir: true,
		local: "static/templates/client",
	},
}

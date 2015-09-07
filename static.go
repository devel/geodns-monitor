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
		local: "static/js/dns.js", size: 2430, modtime: 1379711203,
		compressed: `
H4sIAAAJbogA/5xWzY7bNhC++ykG7CKiUkdeF9vLOtkCbS/toWnRYxAYXGksMZFEhqR2axh+9w4l0pYU
LRqUF9Hz+80vzTqLYJ2RuWO71YofujZ3UrXAb1I4rVZA50kYyDtjsHV7rbR6QrO7cjpdCIfwDqIq94oQ
zk1Wovv97/d/8GQjtNxYJ1xnk/VVeqCMdfzZbHLVWlVjVquSs5ytIQjuJnIegUVDkCxBGESyQNjNLMI+
O8jaoeFTuTEYwgEGXWdaIP4g1giXV5xts7tblsI5XU3slkboiqJs0VAeflG1Mpa/f/yEucs+49Hy4CXN
amxLV6W7qf4NT76LEVgt2g8G63dOqdpJ/TFJs3DlSSULTGbhj5XdoyqOpFC5puaMzST3GQqKgr5WGffz
kf9X8K1okIKd8Sc2r+EbzJUp+KBFpcq+6Hml/Bn4vlJ7yqoO8mMM7QjDCTz3Hlo4E5Ala7nPNlmLNXB9
+oPZRQ3Ctc9rYftu8b/g1atweYDtj7fwE7BKltWbLx2a4xtfUgb3wNgLxrZNNLRtqFR/UiKk9YHcLbo3
aDX1Ne6dbPAChM8YA6Yp6QHubm9TD8/W6vlNZL4Ezs+Fw0bXw2zGa+x5Mt4WNAinMDz3YBdTPOov6iyh
NanxaGwmf553Nk3cpDuY7ZpGmONllrNAWOgW0u19DwKxq8dxBNUQSLAIKfzf+TrREixLnwpW+RVHKMlV
jg2tPaIRs3Lsqxj7/eN8D/oll/lN2pbycAwrZg0dwTvIFos1/LA0vL3YvugaHYMkGyPBWJVzcBwSU+Bj
V+6dKsualkJGDZfktcw/jzcrxnklnWV/gz4PLkjsEt+08GPzTGRSsyU3/mCmDT5Ryn7Fg+hqx0eh+FxJ
Tam64a6StBAd/jMRmK786YtDHmeUFxU9UCyAHrEeqNRr6P3BVUEegM/N0b6ZkbLwjbsXztd3LYQQJU5w
bZ9GtJ2oyTNz0tXIiHRQik0GbK7PbKWexzt7BuWatVCgS6U2r+cF69ucipSM4I1T5fqOXnyuv6lM/oQV
nbylcj8kX+2N75fHIb64H2j0ZqOxYCJ5u+mtj4bhErr/vt5cRiJmqcBGhXQw0U964Hxk0cMlJxdC3zD8
W/s5sPzfgP46/P+J+XmWbaGeKU73G+XZPImaDwJr2G5pge9W55R/+ss/LnT/NwAA//881WEjfgkAAA==
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
		local: "static/js/templates.js", size: 2062, modtime: 1379711203,
		compressed: `
H4sIAAAJbogA/7xVX2/bIBB/3j4F4mGyNdeu463NH8fP+wB7jGQRhyRoNlDAmao0332HSdOG2FqySouE
Aj74/e53cHc7opChjayJoRrN0f4w+3xax5qqHVXwmdPf6IfYEB7/PBqDdcsrwwQPqkhGLNxbpHJutkzP
yngZsDl7ecE47BY4N6p4nS84Rl8RC2dsHZSxhrEOsGPCkQVLw+7vPkofom+TUYT3e3Q44DDcl7HSHd8Z
eWkNHcmqyLUkHGnzXNP5Ai9J9WujRMtXd5WohZqqDexzfpTxzlF3Fsd8H4ZHJ8MFLr7wpZazPLGQBXLI
itaAa4SoDZMLjAwzHRM+l8NJQ7WvJh1H6ejxWjkn/yyW7x4CvkM4q2IpZABTF1hc+Nr6zh4F5QmE6+JO
elZvYa1qojWItcoHjuYEbRVdw6atMXKaJL5HTPr+TMf337NEG2Ja3SPh8kCekCH2GzSd1PiET1KXnc3n
HRSN4Oc9AADxrn/0+BCNxrdf/wnpzZHEfut7AP8cjIF9ny5VpY2nK0snUZZdnaXvdb1i/Q9lJ14oMxp8
unxUg5guARpS14UXD1tbLi46m2QwJrcFZBXg+Kokh+ztPPmLv77qVhrW0LInl26CgbQwZStX0AA+iOTy
/QYQWPWGw3UWRU2rOALoGmx23/s21jYNUc8f62Nn3coB+gV+FI3HV3crWzTOy+rScATjjvG16CY1URva
UxJ7SgJ6aqli0MIl9GtNK8FXxzqPh+LXE7I/AQAA//9tTrLhDggAAA==
`,
	},

	"/templates/client/server.html": {
		local: "templates/client/server.html", size: 510, modtime: 1379711203,
		compressed: `
H4sIAAAJbogA/1RR0U4DIRB8lq8gNDH6oGiMiakcv9LQFtuL3IGwZ2I2/Lu7R8+efegtw+zsMGsgW4G4
KT5/+1yrMHC0piQ3ygI/wXdq7w6fpxyn8fhwiCHmbT7t7xDnutZ7ZW/HfUnvRnOTla03+9ApiDFAn5SE
HliJxoxu8KVWRC5qlYj6Aim7gBclo8mJWPk5BFdKp0jQCuPkOfuPTp0B0lZrxD7Vun17en3RBRxMhfUY
M9oR/U9rUUH8SmU3H3i2kPQjfwSyu/mj6Z/8zfVKwYqbRnweLlQurmQ+remIFGzp48heGG0vGlwIdLfh
aNvQx5bHAlAMM+fahDgl6Ae/S2spRHoE7KZ0dOD/X7QkFozWrK9r1rz43wAAAP//Fuc+a/4BAAA=
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
		local: "templates/index.html", size: 3650, modtime: 1379711203,
		compressed: `
H4sIAAAJbogA/6RXbY8TNxD+zq8Ytt8Qu3u5AgfHJl8KAiSq0kIrVQhVznqy6zuvbWxvjrRqf3vH9ia7
R8JdTpyUnF+eeTwez1uq+y9++enDn+9eQus7ubhXhX8gmWrmGapscQ+gapHxMKBhh55B3TLr0M+z3q/y
p9mw5YWXuGhQc+Wg00p4basyrU6EFetwnq0FXhltfQa1Vh4VkV0J7ts5x7WoMY+ThyCIRTCZu5pJnM+K
k2yfiqOrrTBeaDVhOwBkvW+1vY5JoPt5Dm8RnN9IdJDng6wU6hJai6t51npvzsvS+cKIrikU+rLmqpRi
6cql1t55y0x5WtZuMi86oQpaycCinGeJvkX0W+XiShoDLDXfwD/DBMAwzoVqcq/NOTw5MV+eQ/kgDsBr
6Nglgm8xXoYJhRYaDUzKuHjFNgEUhkvtve5Ar+KMyJbMwoNyOObfpEc5UeS7b51bdEYrJ9Z4kwH2DqNj
mBd1JCMXullueLHXH35++xhcK7qHsNIW3rx8kj8F15vgW+HSCYASO3rxydOS/EexAulJBJ592r5BlVwJ
nK1JoTLEwuPAXjRaNxJrzbGodVe6tSq97dVlghQXLluQFaPw9oSPqLhYfQpHXtN5xdbAFKe36OsWBD3g
1y6XLk2+6uveR0Q22Eh0rEFXEkVYLegr2xNkxkjMI3t+QPbr7cKo5hgSJ/5GN8/OTr+cnd5CmUfQHYln
s0df6HMb9QDbkVdlyk5hGCJoOJCLNdSSOSK+Ip80aHeuE7YEn2cdhU22Be3CKMCqXm7XFT0WfXLPli6d
J8WiYoOSP7S6www484x0bMhD5hkhs8VrWq9KRj5B8D2phjRqD4m9ChvflnMb57FzhyTfp61vy3Jc9s0h
yRdhY5Sryp6qANlgYkHC5UPSzPbtG3YNUwis9hTzWTRuNMxg8YnNXd91zG5CsNDSuE8cEieENInf4ViO
yiEf5pRhhEGeDnFo12hdtoteP1aqMLOLXTYNU54S/FBpzuGMcilZLpJQoeI3o2enEf7m3e3QpxH5a49W
oDsW/t+MkiV8NkcIJMX/oKtTzTuW/5XV/THkCf0bUja2Hvmx2vxuyLWOgc9OIn4KpLGdjK8/YgrqcXcM
8jgNXrGN7IlPHXLQ5DQp+EYdK7NVcEUunodsdO46qqTPIS2EzXPhmRT1EKEOmEXgKNmGHHO5AU3e7zD4
alGVZsJdM7Vmbjz3/eCxkJqd7NnJCSU8FE1L7cgjmpBdkswNJB80KXOYYrZPsbXKLWbZ5pbx3CGlUA9G
lbVjob8q7mDnlHB2L2XIYvEcKvK9+4v3nRnPeqtZ6HWKotg9LOEP6R8nqZSWY+sTa+z1KnvByAKxHsfl
+Z3/rvUK8E6ymt6a+dhGUWXfdlRc133oLcClfsuEsgWSLkSFngxoR55pb3FTZ3XxmTLHppwVZ8VsmMRG
6kCTcSzltFm7+LpD/Q7enrKzdbW2SOr+SOqOC/k+dbTkndRm9eWSgovInxH5drpPfXcjt7phinhPisdp
fIwphg6VLEiRYSQlPPc91nMdvUMr6Hq74QGr3aBFraW2t6ow4mP6OB4euvDr4NhwpQxclekH473/AwAA
//+91VQPQg4AAA==
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

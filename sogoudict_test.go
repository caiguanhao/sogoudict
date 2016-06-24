package sogoudict_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/caiguanhao/sogoudict"
)

func Test_ErrInvalidDict(t *testing.T) {
	_, err := sogoudict.Parse(&bytes.Reader{})
	if err != sogoudict.ErrInvalidDict {
		t.Error("error should be sogoudict.ErrInvalidDict")
	}
}

func Example_parseFile() {
	dict, err := sogoudict.ParseFile("test/fixtures/programming.scel")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, item := range dict.Items {
		fmt.Println(item.Text, strings.Join(item.Abbr, ""), item.Pinyin)
	}
	// Output:
	// 哈希 hx [ha xi]
	// 第一类对象 dyldx [di yi lei dui xiang]
	// 方法 ff [fang fa]
	// 初始化 csh [chu shi hua]
	// 伪变量 wbl [wei bian liang]
	// 全局变量 qjbl [quan ju bian liang]
	// 局部变量 jbbl [ju bu bian liang]
	// 实例变量 slbl [shi li bian liang]
	// 类变量 lbl [lei bian liang]
	// 变量 bl [bian liang]
	// 常量 cl [chang liang]
	// 析构函数 xghs [xi gou han shu]
	// 构造函数 gzhs [gou zao han shu]
	// 访问器 fwq [fang wen qi]
	// 属性 sx [shu xing]
	// 成员方法 cyff [cheng yuan fang fa]
	// 成员函数 cyhs [cheng yuan han shu]
	// 成员属性 cysx [cheng yuan shu xing]
	// 成员 cy [cheng yuan]
	// 实例 sl [shi li]
	// 函数式 hss [han shu shi]
	// 面向过程 mxgc [mian xiang guo cheng]
	// 面向对象 mxdx [mian xiang dui xiang]
}

func Example_parseHTTP() {
	// http://pinyin.sogou.com/dict/detail/index/33688
	url := "http://download.pinyin.sogou.com/dict/download_cell.php?id=33688&name="
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Referer", "http://pinyin.sogou.com/")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	dict, err := sogoudict.Parse(bytes.NewReader(body))
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, item := range dict.Items {
		fmt.Println(item.Text, strings.Join(item.Abbr, ""), item.Pinyin)
	}
	// Output:
	// 春鸽 cg [chun ge]
	// 法克鱿 fky [fa ke you]
	// 雅麽蝶 ymd [ya mo die]
	// 菊花蚕 jhc [ju hua can]
	// 吟稻燕 ydy [yin dao yan]
	// 吉跋猫 jbm [ji ba mao]
	// 达菲鸡 dfj [da fei ji]
	// 潜烈蟹 qlx [qian lie xie]
	// 尾申鲸 wsj [wei shen jing]
	// 草泥马 cnm [cao ni ma]
}

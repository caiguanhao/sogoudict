package sogoudict_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
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
		fmt.Println(item.Text, item.Pinyin)
	}
	// Output:
	// 哈希 [ha xi]
	// 第一类对象 [di yi lei dui xiang]
	// 方法 [fang fa]
	// 初始化 [chu shi hua]
	// 伪变量 [wei bian liang]
	// 全局变量 [quan ju bian liang]
	// 局部变量 [ju bu bian liang]
	// 实例变量 [shi li bian liang]
	// 类变量 [lei bian liang]
	// 变量 [bian liang]
	// 常量 [chang liang]
	// 析构函数 [xi gou han shu]
	// 构造函数 [gou zao han shu]
	// 访问器 [fang wen qi]
	// 属性 [shu xing]
	// 成员方法 [cheng yuan fang fa]
	// 成员函数 [cheng yuan han shu]
	// 成员属性 [cheng yuan shu xing]
	// 成员 [cheng yuan]
	// 实例 [shi li]
	// 函数式 [han shu shi]
	// 面向过程 [mian xiang guo cheng]
	// 面向对象 [mian xiang dui xiang]
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
		fmt.Println(item.Text, item.Pinyin)
	}
	// Output:
	// 春鸽 [chun ge]
	// 法克鱿 [fa ke you]
	// 雅麽蝶 [ya mo die]
	// 菊花蚕 [ju hua can]
	// 吟稻燕 [yin dao yan]
	// 吉跋猫 [ji ba mao]
	// 达菲鸡 [da fei ji]
	// 潜烈蟹 [qian lie xie]
	// 尾申鲸 [wei shen jing]
	// 草泥马 [cao ni ma]
}

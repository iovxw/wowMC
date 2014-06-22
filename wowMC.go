/*
__/\\\\\\\\\\\\\____/\\\\\\__________________________________________________________/\\\________/\\\\\\\_______________/\\\____
 _\/\\\/////////\\\_\////\\\__________________________________/\\\__________________/\\\\\______/\\\/////\\\___________/\\\\\____
  _\/\\\_______\/\\\____\/\\\_________________________________\/\\\________________/\\\/\\\_____/\\\____\//\\\________/\\\/\\\____
   _\/\\\\\\\\\\\\\\_____\/\\\_____/\\\____/\\\_____/\\\\\\\\__\/\\\\\\\\_________/\\\/\/\\\____\/\\\_____\/\\\______/\\\/\/\\\____
    _\/\\\/////////\\\____\/\\\____\/\\\___\/\\\___/\\\/////\\\_\/\\\////\\\_____/\\\/__\/\\\____\/\\\_____\/\\\____/\\\/__\/\\\____
     _\/\\\_______\/\\\____\/\\\____\/\\\___\/\\\__/\\\\\\\\\\\__\/\\\\\\\\/____/\\\\\\\\\\\\\\\\_\/\\\_____\/\\\__/\\\\\\\\\\\\\\\\_
      _\/\\\_______\/\\\____\/\\\____\/\\\___\/\\\_\//\\///////___\/\\\///\\\___\///////////\\\//__\//\\\____/\\\__\///////////\\\//__
       _\/\\\\\\\\\\\\\/___/\\\\\\\\\_\//\\\\\\\\\___\//\\\\\\\\\\_\/\\\_\///\\\___________\/\\\_____\///\\\\\\\/_____________\/\\\____
        _\/////////////____\/////////___\/////////_____\//////////__\///____\///____________\///________\///////_______________\///_____
*/
package main

import (
	"encoding/json"
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/salviati/go-qt5/qt5"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var (
	exit         = make(chan bool)
	alreadyCheck bool
	config       Config
)

func main() {
	if _, err := os.Stat("config.json"); err == nil { //检查配置文件是否存在
		bi, _ := ioutil.ReadFile("config.json") //读配置文件
		json.Unmarshal(bi, &config)
	}
	qt5.Main(func() {
		go ui_main()
		qt5.Run()
		exit <- true
	})
}

func ui_main() {
	w := qt5.NewWidget()
	w.SetWindowTitle("登陆窗口") //设置窗口标题
	w.SetSizev(300, 100)     //设置窗口大小

	//================= vbox1 =================

	vbox1 := qt5.NewVBoxLayout() //创建一个容器

	lbl1 := qt5.NewLabel() //创建标签
	lbl1.SetText("用户名：")   //设置文本

	lbl2 := qt5.NewLabel() //创建标签
	lbl2.SetText("密码：")    //设置文本

	vbox1.AddWidget(lbl1) //在容器里添加控件
	vbox1.AddWidget(lbl2) //在容器里添加控件

	//================= vbox2 =================

	vbox2 := qt5.NewVBoxLayout()

	ed1 := qt5.NewLineEdit() //创建编辑框
	if config.Name != "" {
		ed1.SetText(config.Name)
	}

	ed2 := qt5.NewLineEdit()                                           //创建编辑框
	if config.Chk1 == 2 && config.Chk2 != 2 && config.Password != "" { //如果开启保存密码，且不保存密匙，且保存的密码不为空，那么读取
		ed2.SetText(config.Password)
	}
	ed2.SetEchoMode(2) //设置文本框为密码模式

	vbox2.AddWidget(ed1) //在容器里添加控件
	vbox2.AddWidget(ed2) //在容器里添加控件

	//================= hbox1 =================

	hbox1 := qt5.NewHBoxLayout()

	hbox1.AddLayout(vbox1)
	hbox1.AddLayout(vbox2)

	//================= hbox2 =================

	hbox2 := qt5.NewHBoxLayout()

	chk2 := qt5.NewCheckBox() //创建一个选择框
	chk2.SetText("使用密匙")
	if config.Chk2 == 2 {
		chk2.SetCheck(2)
	} else {
		config.Play.ClientToken = "" //如果不使用密匙，那么清空密匙
	}

	chk1 := qt5.NewCheckBox() //创建一个选择框
	chk1.SetText("记住密码")
	if config.Chk1 == 2 {
		chk1.SetCheck(2)
	} else {
		chk2.Hide() //如果没有被选定，那么不显示附加选项
	}

	chk3 := qt5.NewCheckBox() //创建一个选择框
	chk3.SetText("自动登录")
	if config.Chk3 == 2 {
		chk3.SetCheck(2)
		if config.Chk1 != 2 { //如果自动登陆被选定，但是记住密码没有被选定，那么选定
			chk1.SetCheck(2)
		}
	}

	hbox2.AddWidget(chk1) //在容器里添加控件
	hbox2.AddWidget(chk2) //在容器里添加控件
	hbox2.AddWidget(chk3) //在容器里添加控件

	//================= vbox3 =================

	vbox3 := qt5.NewVBoxLayout()

	btn1 := qt5.NewButtonWithText("确定") //直接用文本为内容创建一个按钮（其他控件也可以这么创建）

	vbox3.AddLayout(hbox1) //在容器里添加控件
	vbox3.AddLayout(hbox2) //在容器里添加控件
	vbox3.AddWidget(btn1)  //在容器里添加控件

	//================= BOX END =================

	btn1.OnClicked(func() { //按钮1被点击时触发的事件
		var (
			accessToken, clientToken, id, name string
			badlogin                           bool
		)
		fmt.Println("用户名：", ed1.Text())
		fmt.Println("密码：", ed2.Text())

		if chk2.Check() == 2 && config.Play.ClientToken != "" {
			accessToken, clientToken, id, name, badlogin = refresh(config.Play.AccessToken, config.Play.ClientToken, config.Play.SelectedProfile.ID, config.Play.SelectedProfile.Name)
		} else {
			name, id, accessToken, clientToken, badlogin = authenticate(ed1.Text(), ed2.Text(), "")

			if badlogin {
				return
			}
			fmt.Println("Name："+name, "ID："+id, "Access Token："+accessToken, "Client Token："+clientToken)
		}

		if chk1.Check() == 2 { //判断是否要保存密码
			if chk2.Check() == 2 { //判断是否要保存密匙
				config = Config{ed1.Text(), "", chk1.Check(), chk2.Check(), chk3.Check(), Refresh{accessToken, clientToken, SelectedProfile_type{id, name}}}
			} else { //如果没有开启保存密匙，那么只保存用户名密码和选项
				config = Config{ed1.Text(), ed2.Text(), chk1.Check(), chk2.Check(), chk3.Check(), Refresh{"", "", SelectedProfile_type{"", ""}}}
			}

		} else { //如果没有开启保存密码，那么只保存用户名和选项
			config = Config{ed1.Text(), "", chk1.Check(), chk2.Check(), chk3.Check(), Refresh{"", "", SelectedProfile_type{"", ""}}}
		}

		b, _ := json.Marshal(config)             //格式化json
		ioutil.WriteFile("config.json", b, 0644) //保存数据到文件

		go luanch()

		//qt5.Exit(0)
	})

	ed1.OnTextChanged(func(string) { //当编辑框1内文本被改变时触发的事件
		fmt.Println(ed1.Text())
	})

	ed2.OnTextChanged(func(string) { //当编辑框2内文本被改变时触发的事件
		fmt.Println(ed2.Text())
	})

	chk1.OnStateChanged(func(int) { //当选择框选择被改变时触发的事件
		fmt.Println("保存密码选项改变")
		if chk1.Check() == 0 {
			chk3.SetCheck(0)
			chk2.Hide()
		} else {
			chk2.Show()
		}
	})

	chk2.OnStateChanged(func(int) { //当选择框选择被改变时触发的事件
		fmt.Println("密匙替换选项被改变")
	})

	chk3.OnStateChanged(func(int) { //当选择框选择被改变时触发的事件
		fmt.Println("自动登陆选项改变")
		if chk3.Check() == 2 {
			if chk1.Check() == 2 { //检查保存密码选项是否已经被选择，保存状态
				alreadyCheck = true
			} else {
				alreadyCheck = false
				chk1.SetCheck(2) //如果保存密码选项没有被开启，那么开启保存密码
			}
			fmt.Println("自动登陆开启")
		} else {
			if alreadyCheck {
				//如果保存密码选项之前就被开启了，那么不做改变（继续保持开启）
			} else {
				chk1.SetCheck(0) //保存密码选项之前没有被开启，那么恢复原来的状态
			}
			fmt.Println("自动登陆关闭")
		}
	})

	w.SetLayout(vbox3)

	defer w.Close()
	w.Show()
	<-exit
}

func authenticate(name_l, password, clientToken_l string) (name, id, accessToken, clientToken string, badlogin bool) {
	x := Login{Agen{"Minecraft", 1}, name_l, password, clientToken_l}
	y, _ := json.Marshal(x) //格式化json
	fmt.Println("POST：", string(y))

	url := "https://authserver.mojang.com/authenticate"
	b := strings.NewReader(string(y))
	resp, err := http.Post(url, "application/json", b)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(resp)

	body, _ := ioutil.ReadAll(resp.Body) //读取返回的json
	//fmt.Println(string(body))

	js, err := simplejson.NewJson(body) //解析json
	if err != nil {
		panic(err.Error())
	}

	js_map, _ := js.Map() //把解析的json提取到map里

	if resp.Status != "200 OK" {
		fmt.Println("验证失败")
		badlogin = true
		fmt.Println(js_map) //如果验证出错，那么打印错误信息（这里就懒得给用户解释详细错误了）
		return
	} else {
		fmt.Println("验证成功")
	}

	accessToken = js_map["accessToken"].(string)
	//fmt.Println(accessToken)
	clientToken = js_map["clientToken"].(string)
	//fmt.Println(clientToken)
	//fmt.Println(js_map["selectedProfile"])
	//fmt.Println(js_map["availableProfiles"])
	name = js_map["selectedProfile"].(map[string]interface{})["name"].(string)
	//fmt.Println(name)
	id = js_map["selectedProfile"].(map[string]interface{})["id"].(string)
	//fmt.Println(id)
	//fmt.Println(js_map["availableProfiles"].([]interface{})[0].(map[string]interface{})["id"]) //这里的数组是多账号切换用的，暂时先不开发
	//fmt.Println(js_map["availableProfiles"].([]interface{})[0].(map[string]interface{})["name"])
	return
}

func refresh(accessToken_r, clientToken_r, id_r, name_r string) (accessToken, clientToken, id, name string, badlogin bool) {
	x := Refresh{accessToken_r, clientToken_r, SelectedProfile_type{id_r, name_r}}
	y, _ := json.Marshal(x) //格式化json
	fmt.Println("POST：", string(y))

	url := "https://authserver.mojang.com/refresh"
	b := strings.NewReader(string(y))
	resp, err := http.Post(url, "application/json", b)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(resp)

	body, _ := ioutil.ReadAll(resp.Body) //读取返回的json
	//fmt.Println(string(body))

	js, err := simplejson.NewJson(body) //解析json
	if err != nil {
		panic(err.Error())
	}

	js_map, _ := js.Map() //把解析的json提取到map里

	if resp.Status != "200 OK" {
		fmt.Println("验证失败")
		badlogin = true
		fmt.Println(js_map) //如果验证出错，那么打印错误信息（这里就懒得给用户解释详细错误了）
		return
	} else {
		fmt.Println("验证成功")
	}
	return
}

func luanch() {
	launcher_profiles, _ := ioutil.ReadFile("/home/bluek404/.minecraft/launcher_profiles.json")
	js, err := simplejson.NewJson(launcher_profiles) //解析json
	if err != nil {
		panic(err.Error())
	}

	js_map, _ := js.Map() //把解析的json提取到map里
	//fmt.Println(js_map)
	//js_version := make([]map[string]interface{}, 0, 10)
	for _, value := range js_map["profiles"].(map[string]interface{}) {
		name := value.(map[string]interface{})["name"].(string)
		file := "/home/bluek404/.minecraft/versions/" + name + "/" + name + ".json"
		if _, err := os.Stat(file); err == nil {
			fmt.Println("名称：", name)
			bi, _ := ioutil.ReadFile(file)
			//fmt.Println(string(bi))
			var json_v Json_version
			json.Unmarshal(bi, &json_v)
			//fmt.Println(json_v.Libraries)
			for _, value := range json_v.Libraries {
				//fmt.Println(value)
				name := value["name"].(string)
				//fmt.Println(name)
				x := strings.Index(name, ":")
				t1 := strings.Split(string(name[x-1:len(name)]), ":")
				if value["natives"] != nil {
					natives := (value["natives"].(map[string]interface{}))["linux"]
					if natives != nil {
						t2 := strings.Replace(string(name[0:x]), ".", "/", -1) + "/" + t1[1] + "/" + t1[2] + "/" + t1[1] + "-" + t1[2] + "-" + natives.(string) + ".jar"
						fmt.Println(t2)
					}
				} else {
					t2 := strings.Replace(string(name[0:x]), ".", "/", -1) + "/" + t1[1] + "/" + t1[2] + "/" + t1[1] + "-" + t1[2] + ".jar"
					fmt.Println(t2)
				}
			}
		}

	}
}

type Config struct { //用于配置文件的type
	Name     string
	Password string
	Chk1     int
	Chk2     int
	Chk3     int
	Play     Refresh
}

type Login struct {
	Agent       Agen   `json:"agent"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	ClientToken string `json:"clientToekn"`
}
type Agen struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}
type Refresh struct {
	AccessToken     string               `json:"accessToken"`
	ClientToken     string               `json:"clientToken"`
	SelectedProfile SelectedProfile_type `json:"selectedProfile"`
}
type SelectedProfile_type struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Json_version struct {
	ID                     string                   `json:"id"`
	Time                   string                   `json:"time"`
	ReleaseTime            string                   `json:"releaseTime"`
	Type                   string                   `json:"type"`
	MinecraftArguments     string                   `json:"minecraftArguments"`
	Libraries              []map[string]interface{} `json:"libraries"`
	MainClass              string                   `json:"mainClass"`
	MinimumLauncherVersion int                      `json:"minimumLauncherVersion"`
	Assets                 string                   `json:"assets"`
}

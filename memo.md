```	
    quitSeq, _ := keymap.ParseKeySequence("C-x C-c")
	e.keymap.Bind(quitSeq, "quit")
```
が設定者側としてはちょっとだるいかも e.BindKey("C-x C-c", "quit") くらいの使い方ができると良さそう
(今後メジャーモードとかマイナーモードのキーマップを扱う必要がある)

C-x とか C-c とかその後にもキーを受け取るものは特殊なものとして扱う必要がありそう

再描画で画面がちらつく

forward-char とかはパッケージ開発者にも呼び出せるようにするべきだけど、e.forwardChar 自身は提供すべきではないと思う この辺のAPI設計というか公開・非公開の仕組みは必要

{{$map := LVMapPack . "Title" "模板主页" "Array" "value1,value2,value3"}}

{{LVEmbedTempate "/base/head.tpl" $map}}

<h1>{{.Content}}</h1>
<div>注意标题名称了没有？</div>

{{LVEmbedTempate "/base/foot.tpl" .}}
extends RefCounted

const CONNECT_TIMEOUT_MS := 3000
const REQUEST_TIMEOUT_MS := 8000

var base_url := ""
var last_error := ""
var _host := ""
var _port := 80
var _client := HTTPClient.new()


func _init(url := "http://127.0.0.1:8080") -> void:
	base_url = url
	var rest := url
	if rest.begins_with("http://"):
		rest = rest.substr(7)
	elif rest.begins_with("https://"):
		last_error = "暂不支持 https 服务地址"
		rest = rest.substr(8)
	rest = rest.split("/")[0]
	var parts := rest.split(":")
	_host = parts[0]
	if parts.size() > 1:
		_port = int(parts[1])


func get_json(path: String) -> Variant:
	return _request(HTTPClient.METHOD_GET, path, "")


func post_json(path: String, body: Variant) -> Variant:
	return _request(HTTPClient.METHOD_POST, path, JSON.stringify(body))


func _ensure_connected() -> bool:
	if _client.get_status() == HTTPClient.STATUS_CONNECTED:
		return true
	_client.close()
	if _host.is_empty():
		last_error = "服务地址未配置"
		return false
	if _client.connect_to_host(_host, _port) != OK:
		last_error = "无法连接 %s:%d" % [_host, _port]
		return false
	var deadline := Time.get_ticks_msec() + CONNECT_TIMEOUT_MS
	while _client.get_status() == HTTPClient.STATUS_CONNECTING or _client.get_status() == HTTPClient.STATUS_RESOLVING:
		_client.poll()
		if Time.get_ticks_msec() > deadline:
			_client.close()
			last_error = "连接超时 %s:%d" % [_host, _port]
			return false
		OS.delay_msec(1)
	if _client.get_status() != HTTPClient.STATUS_CONNECTED:
		last_error = "连接失败 %s:%d" % [_host, _port]
		return false
	return true


func _request(method: HTTPClient.Method, path: String, body: String) -> Variant:
	last_error = ""
	if not _ensure_connected():
		return null
	var headers := PackedStringArray(["Content-Type: application/json", "Accept: application/json"])
	if _client.request(method, path, headers, body) != OK:
		last_error = "请求发送失败"
		_client.close()
		return null
	var deadline := Time.get_ticks_msec() + REQUEST_TIMEOUT_MS
	while _client.get_status() == HTTPClient.STATUS_REQUESTING:
		_client.poll()
		if Time.get_ticks_msec() > deadline:
			_client.close()
			last_error = "请求超时"
			return null
		OS.delay_msec(1)
	if not _client.has_response():
		last_error = "无响应"
		_client.close()
		return null
	var code := _client.get_response_code()
	var raw := PackedByteArray()
	while _client.get_status() == HTTPClient.STATUS_BODY:
		_client.poll()
		var chunk := _client.read_response_body_chunk()
		if chunk.size() == 0:
			if Time.get_ticks_msec() > deadline:
				_client.close()
				last_error = "读取响应超时"
				return null
			OS.delay_msec(1)
		else:
			raw.append_array(chunk)
	var parsed: Variant = JSON.parse_string(raw.get_string_from_utf8())
	if code >= 400:
		if parsed is Dictionary and parsed.has("error"):
			last_error = str(parsed["error"])
		else:
			last_error = "HTTP %d" % code
		return null
	if parsed == null:
		last_error = "响应不是合法 JSON"
	return parsed

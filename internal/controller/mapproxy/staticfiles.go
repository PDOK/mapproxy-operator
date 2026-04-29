package mapproxy

var templateInclude = `
server.modules += ( "mod_status" )

$HTTP["remoteip"] =~ "^(127\.0\.0\.1|172\.(1[6-9]|2[0-9]|3[01])\.|10\.|192\.168\.)" {
    status.status-url = "/server-status"
}

url.rewrite-once = (
$REWRITERULES
)

magnet.attract-response-start-to = ( "/srv/mapproxy/config/response.lua" )
`

var responseLua = `
if lighty.r.req_item.http_status == 500 then
    if string.find(lighty.r.resp_body.get, "invalid request") then
        body = string.gsub(lighty.r.resp_body.get, "internal error: ", "")
        lighty.r.resp_body.set(body)
        lighty.r.resp_header["Content-Length"] = string.len(body)
        return 400
    end
end
`

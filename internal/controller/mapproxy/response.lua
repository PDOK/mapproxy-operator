if lighty.r.req_item.http_status == 500 then
    if string.find(lighty.r.resp_body.get, "invalid request") then
        body = string.gsub(lighty.r.resp_body.get, "internal error: ", "")
        lighty.r.resp_body.set(body)
        lighty.r.resp_header["Content-Length"] = string.len(body)
        return 400
    end
end
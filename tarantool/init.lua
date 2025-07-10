box.cfg{
    listen = 3301
}

box.once("kv_space", function()
    local s = box.schema.space.create("kv")
    s:format({
        { name = 'key', type = 'string' },
        { name = 'value', type = 'string' },
    })
    s:create_index('primary', { parts = { 'key' } })
end)

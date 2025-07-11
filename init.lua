box.cfg{
    listen = 3301
}

-- Создаем спейс, если его нет
if not box.space.kv then
    s = box.schema.space.create('kv')
    s:format({
        {name = 'key', type = 'string'},
        {name = 'value', type = 'string'},
    })
    s:create_index('primary', {parts = {'key'}})
end

-- Гостю выдаем права на чтение и запись
box.schema.user.grant('guest', 'read', 'space', 'kv', {if_not_exists=true})
box.schema.user.grant('guest', 'write', 'space', 'kv', {if_not_exists=true})
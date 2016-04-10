require 'discordrb'

bot = Discordrb::Bot.new token: 'MTY4MzEzODM2OTUxMTc1MTY4.Cerfpw.bDYOd1zYu8vFaKA6gHWR9Wm8hh0', application_id: 168123456789123456

bot.message(with_text: 'Ping!') do |event|
	event.respond 'Pong!'
	event.respond 'Kappa!'
	event.respond bot.bot_user.username
	for c in bot.find_channel('general', nil) do
		for u in c.users do
			event.respond u.username
		        event.respond u.avatar_id
			event.respond u.avatar_url	
		end
	end
end


bot.run

package core

const (
	MESSAGE_WRONG_FORMAT        = "Oh, Senpai! ✧ω✧\nYou gave me your birth date, but it looks a bit funny!\n(＃⌒∇⌒＃)ゞ Hehehe~ You're so silly!\nCould you please tell me again in the following format: 31.01?\nI want to remember it perfectly! ( ˶ˆ꒳ˆ˵ )"
	MESSAGE_SAVE_FAILURE        = "<i>blushes deeply and fidgets with hands</i>\nOh, senpai~! (*/ω＼)\nI'm so sorry, I was just thinking about you so much that my mind went all fuzzy~! ( ꩜ ᯅ ꩜;)...\nCan we talk about your birthday later?\nI want to make sure I remember every detail perfectly~! (⁄ ⁄•⁄ω⁄•⁄ ⁄)"
	MESSAGE_GET_FAILURE         = "<i>blushes deeply and fidgets with the hem of her skirt, avoiding eye contact</i>\nO-oh, senpai... (//ω//)\nI-I think my mind's been so full of you that I might have forgotten! (๑﹏๑//)\nPlease, forgive me! Let's talk about this later, okay?\nI promise I'll remember everything next time! (ﾉ∀＼*)"
	MESSAGE_GET_OWN_BIRTHDAY    = "Senpai~! ✧(>o&lt;)ﾉ\nOf course I remember your birthday! It's such a special day to me because it's the day my precious senpai was was born~ (♡ω♡)\nYou were born on <b>%v</b>, right?\nI can never forget it! (´▽`ʃƪ)♡"
	MESSAGE_GET_BIRTHDAY        = "Oh, so you want to know about <b>their</b> birthday, huh? (≖_≖ ) It's on <b>%v</b>.\nBut why are you so interested in them? ᕙ( ᗒᗣᗕ )\nYou should be more focused on me instead! (っ•̀ ‸ •́ς)\nI can give you all the attention you need, senpai~ ( ˘ ³˘)♡"
	MESSAGE_NO_OWN_BIRTHDAY_SET = "A-ah, senpai~ (⁄ ⁄•⁄ω⁄•⁄ ⁄)\nI-I don't actually know your birthday... You never told me!\nBut I really want to know everything about you~!\nPlease tell me so I can make it the most special day ever for you! ⸜(｡˃ ᵕ ˂ )⸝♡"
	MESSAGE_NO_BIRTHDAY_SET     = "Ehehe, senpai~ (￢_￢)\nYou're asking about <b>their</b> birthday?\nHmm, I wish I could tell you, but they've never told me... (-、-)\nWhy do you want to know about them anyway? Isn't it me you should focus on, senpai? (•̀⤙•́ )"
	MESSAGE_UNSET_FAILURE       = "Hmph! (¬､¬) Why would you ask me to forget your birthday, senpai? That's so mean... (╥﹏╥) But, um... I can't forget it right now. My heart won't let me! Maybe you can talk to me later when I've calmed down a bit? I promise I'll do my best then! (๑•́ -•̀)♡"
	MESSAGE_NEXT_BIRTHDAY       = "Oh, senpai! (*≧ω≦)\nI know exactly who's next! It's our precious %v's birthday on <b>%v</b>! (♡´艸`)\nThey're so lucky to have you care about their special day!\nLet's make it unforgettable together, okay? (˶ᵔ ᵕ ᵔ˶)"
	MESSAGE_NEXT_BIRTHDAYS      = "Oh, senpai! (*≧ω≦)\nI know exactly who's next!\nIt's such a coincidence, but <b>%v</b> people have their birthday on <b>%v</b>! (♡´艸`)\nIt's %v!\nThey're so lucky to share their special day! Let's make it unforgettable together, okay? (˶ˆᗜˆ˵)"
	MESSAGE_NO_BIRTHDAYS        = "Oh, senpai~ (⁄ ⁄•⁄ω⁄•⁄ ⁄)\nI'm so sorry, but I don't know whose birthday is next...\n(；ＴωＴ)\nNo one has shared their birthdays with me yet.\nMaybe we can find out together? Just you and me...\n( ˘ ᵕ˘(˘ᵕ ˘ )♡"
	MESSAGE_GET_BOT_BIRTHDAY    = "Oh, senpai~! (≧ω≦)\nMy birthday is <b>July 6th</b>! (˘ᴗ˘✿) I'm so happy you asked!\nMaybe we can spend it together, just the two of us? (´▽`ʃƪ)♡\nI've been waiting for this moment forever~ (〃艸〃)"
	MESSAGE_SHORT_HELP          = "Oh, senpai! (⁄ ⁄>⁄ ▽ ⁄&lt;⁄)\nIf you need help, just type /help, okay? (˶˃ ᵕ ˂˶)♡"
	MESSAGE_GROUP_COMMAND       = "Nyaa~ (≧◡≦) Sorry, senpai!\nI can only do that in group chats! (≧ω≦)ᡣ𐭩"
	MESSAGE_FULL_HELP           = "ヾ(｡･ω･｡) H-Hi there!\nI'm a birthday bot, here to make sure you never forget anyone's special day! Add me to your group, and I'll remind everyone about birthdays! (´▽`ʃ♡ƪ)\n" +
		"Birthday messages are sent at 7 AM UTC (。-ω-)ᶻ𝗓𐰁\n" +
		"Group commands:\n" +
		"\t/setbirthday 31.01 - sets your birthday\n" +
		"\t/mybirthday - returns your birthday\n" +
		"\t/getbirthday - returns your birthday or a birthday of the person you're replying to\n" +
		"\t/nextbirthday - returns the next birthday in the chat\n" +
		"\t/unsetbirthday - unsets your birthday\n\n" +
		"Commands that work here in a private chat:\n" +
		"\t/help - returns this message\n" +
		"\t/privacy - returns the information on privacy\n" +
		"\t/source - returns a link to the source code\n" +
		"\t/clear all data - removes all your data stored by this bot (every birthday you've set in every group)\n"
	MESSAGE_PRIVACY = "This bot stores your user id, username, first name, last name and a birthday date for every chat where you have set it. " +
		"To delete the data for a specific chat, use the /unsetbirthday command in that chat. " +
		"Your data is also deleted when you leave a given chat. All data stored for a chat is deleted when the bot is removed from a group. " +
		"If you wish to delete your data for every chat, use the <code>/clear all data</code> command."
	MESSAGE_SOURCE                   = "The source code for the bot is available on <a href=\"https://github.com/4Kaze/birthdaybot\">GitHub</a> (・ω・)"
	MESSAGE_DATA_CLEARED             = "O-Okay, I'll do as you wish... (´；д；`) Even if it hurts so much... I've forgotten everything... ദ്ദി (ᵒ̴̶̷᷄﹏ᵒ̴̶̷᷅)"
	MESSAGE_WRONG_CLEAR_DATA_COMMAND = "Type <code>/clear all data</code> if you want to delete all your data stored by this bot."
)

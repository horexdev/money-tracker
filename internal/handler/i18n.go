package handler

import "github.com/horexdev/money-tracker/internal/domain"

// botStrings holds all user-visible bot message strings for one language.
type botStrings struct {
	welcome     string
	help        string
	defaultMsg  string
	openButton  string
}

// messages maps each supported language to its localised bot strings.
var messages = map[domain.Language]botStrings{
	domain.LangEN: {
		welcome: "<b>Welcome to MoneyTracker!</b>\n\nYour personal finance companion in Telegram.\nTrack expenses, monitor income, and stay on top of your money.\n\nTap the button below to open the app 👇",
		help:    "<b>MoneyTracker Help</b>\n\nOpen the Mini App to manage your finances:\n• Add expenses and income\n• View balance and transaction history\n• Track budgets and savings goals\n• Set up recurring transactions\n• Export your data\n\nUse the button below to get started.",
		defaultMsg: "Use the button below to open MoneyTracker, or type /help for more info.",
		openButton: "📱 Open MoneyTracker",
	},
	domain.LangRU: {
		welcome: "<b>Добро пожаловать в MoneyTracker!</b>\n\nВаш личный финансовый помощник в Telegram.\nОтслеживайте расходы, контролируйте доходы и управляйте своими финансами.\n\nНажмите кнопку ниже, чтобы открыть приложение 👇",
		help:    "<b>Помощь MoneyTracker</b>\n\nОткройте мини-приложение для управления финансами:\n• Добавляйте расходы и доходы\n• Смотрите баланс и историю транзакций\n• Отслеживайте бюджеты и цели накоплений\n• Настраивайте регулярные транзакции\n• Экспортируйте данные\n\nНажмите кнопку ниже, чтобы начать.",
		defaultMsg: "Нажмите кнопку ниже, чтобы открыть MoneyTracker, или введите /help для получения справки.",
		openButton: "📱 Открыть MoneyTracker",
	},
	domain.LangUK: {
		welcome: "<b>Ласкаво просимо до MoneyTracker!</b>\n\nВаш особистий фінансовий помічник у Telegram.\nВідстежуйте витрати, контролюйте доходи та керуйте своїми фінансами.\n\nНатисніть кнопку нижче, щоб відкрити застосунок 👇",
		help:    "<b>Довідка MoneyTracker</b>\n\nВідкрийте міні-застосунок для керування фінансами:\n• Додавайте витрати та доходи\n• Переглядайте баланс та історію транзакцій\n• Відстежуйте бюджети та цілі накопичень\n• Налаштовуйте регулярні транзакції\n• Експортуйте дані\n\nНатисніть кнопку нижче, щоб почати.",
		defaultMsg: "Натисніть кнопку нижче, щоб відкрити MoneyTracker, або введіть /help для отримання довідки.",
		openButton: "📱 Відкрити MoneyTracker",
	},
	domain.LangBE: {
		welcome: "<b>Сардэчна запрашаем у MoneyTracker!</b>\n\nВаш асабісты фінансавы памочнік у Telegram.\nСачыце за выдаткамі, кантралюйце даходы і кіруйце сваімі фінансамі.\n\nНацісніце кнопку ніжэй, каб адкрыць праграму 👇",
		help:    "<b>Даведка MoneyTracker</b>\n\nАдкрыйце міні-праграму для кіравання фінансамі:\n• Дадавайце выдаткі і даходы\n• Глядзіце баланс і гісторыю транзакцый\n• Сачыце за бюджэтамі і мэтамі зберажэнняў\n• Налайджвайце рэгулярныя транзакцыі\n• Экспартуйце дадзеныя\n\nНацісніце кнопку ніжэй, каб пачаць.",
		defaultMsg: "Націсніце кнопку ніжэй, каб адкрыць MoneyTracker, або ўвядзіце /help для атрымання даведкі.",
		openButton: "📱 Адкрыць MoneyTracker",
	},
	domain.LangKK: {
		welcome: "<b>MoneyTracker-ға қош келдіңіз!</b>\n\nTelegram-дағы жеке қаржы көмекшіңіз.\nШығыстарды қадағалаңыз, табысты бақылаңыз және қаржыңызды басқарыңыз.\n\nҚолданбаны ашу үшін төмендегі түймені басыңыз 👇",
		help:    "<b>MoneyTracker анықтамасы</b>\n\nҚаржыны басқару үшін мини-қолданбаны ашыңыз:\n• Шығыстар мен табыстарды қосыңыз\n• Баланс пен транзакция тарихын қараңыз\n• Бюджеттер мен жинақ мақсаттарын бақылаңыз\n• Тұрақты транзакцияларды реттеңіз\n• Деректерді экспорттаңыз\n\nБастау үшін төмендегі түймені басыңыз.",
		defaultMsg: "MoneyTracker-ды ашу үшін төмендегі түймені басыңыз немесе анықтама алу үшін /help деп жазыңыз.",
		openButton: "📱 MoneyTracker-ды ашу",
	},
	domain.LangUZ: {
		welcome: "<b>MoneyTracker-ga xush kelibsiz!</b>\n\nTelegramdagi shaxsiy moliyaviy yordamchingiz.\nXarajatlarni kuzating, daromadlarni nazorat qiling va moliyangizni boshqaring.\n\nIlovani ochish uchun pastdagi tugmani bosing 👇",
		help:    "<b>MoneyTracker yordami</b>\n\nMoliyani boshqarish uchun mini-ilovani oching:\n• Xarajatlar va daromadlarni qo'shing\n• Balans va tranzaksiyalar tarixini ko'ring\n• Byudjetlar va jamg'arma maqsadlarini kuzating\n• Muntazam tranzaksiyalarni sozlang\n• Ma'lumotlarni eksport qiling\n\nBoshlash uchun pastdagi tugmani bosing.",
		defaultMsg: "MoneyTracker-ni ochish uchun pastdagi tugmani bosing yoki /help yozing.",
		openButton: "📱 MoneyTracker-ni ochish",
	},
	domain.LangES: {
		welcome: "<b>¡Bienvenido a MoneyTracker!</b>\n\nTu asistente financiero personal en Telegram.\nRegistra gastos, controla ingresos y gestiona tu dinero.\n\nToca el botón de abajo para abrir la app 👇",
		help:    "<b>Ayuda de MoneyTracker</b>\n\nAbre la Mini App para gestionar tus finanzas:\n• Añade gastos e ingresos\n• Consulta el saldo e historial de transacciones\n• Controla presupuestos y metas de ahorro\n• Configura transacciones recurrentes\n• Exporta tus datos\n\nUsa el botón de abajo para empezar.",
		defaultMsg: "Usa el botón de abajo para abrir MoneyTracker o escribe /help para obtener ayuda.",
		openButton: "📱 Abrir MoneyTracker",
	},
	domain.LangDE: {
		welcome: "<b>Willkommen bei MoneyTracker!</b>\n\nDein persönlicher Finanzbegleiter in Telegram.\nVerfolge Ausgaben, überwache Einnahmen und behalte dein Geld im Blick.\n\nTippe den Button unten, um die App zu öffnen 👇",
		help:    "<b>MoneyTracker Hilfe</b>\n\nÖffne die Mini App, um deine Finanzen zu verwalten:\n• Ausgaben und Einnahmen hinzufügen\n• Kontostand und Transaktionsverlauf anzeigen\n• Budgets und Sparziele verfolgen\n• Wiederkehrende Transaktionen einrichten\n• Daten exportieren\n\nNutze den Button unten, um loszulegen.",
		defaultMsg: "Nutze den Button unten, um MoneyTracker zu öffnen, oder tippe /help für weitere Infos.",
		openButton: "📱 MoneyTracker öffnen",
	},
	domain.LangIT: {
		welcome: "<b>Benvenuto su MoneyTracker!</b>\n\nIl tuo assistente finanziario personale su Telegram.\nTieni traccia delle spese, monitora le entrate e gestisci il tuo denaro.\n\nPremi il pulsante in basso per aprire l'app 👇",
		help:    "<b>Guida MoneyTracker</b>\n\nApri la Mini App per gestire le tue finanze:\n• Aggiungi spese ed entrate\n• Visualizza saldo e cronologia delle transazioni\n• Monitora budget e obiettivi di risparmio\n• Imposta transazioni ricorrenti\n• Esporta i tuoi dati\n\nUsa il pulsante in basso per iniziare.",
		defaultMsg: "Usa il pulsante in basso per aprire MoneyTracker o digita /help per ulteriori informazioni.",
		openButton: "📱 Apri MoneyTracker",
	},
	domain.LangFR: {
		welcome: "<b>Bienvenue sur MoneyTracker!</b>\n\nVotre assistant financier personnel sur Telegram.\nSuivez vos dépenses, surveillez vos revenus et gérez votre argent.\n\nAppuyez sur le bouton ci-dessous pour ouvrir l'application 👇",
		help:    "<b>Aide MoneyTracker</b>\n\nOuvrez la Mini App pour gérer vos finances:\n• Ajoutez des dépenses et des revenus\n• Consultez le solde et l'historique des transactions\n• Suivez les budgets et les objectifs d'épargne\n• Configurez des transactions récurrentes\n• Exportez vos données\n\nUtilisez le bouton ci-dessous pour commencer.",
		defaultMsg: "Utilisez le bouton ci-dessous pour ouvrir MoneyTracker ou tapez /help pour plus d'informations.",
		openButton: "📱 Ouvrir MoneyTracker",
	},
	domain.LangPT: {
		welcome: "<b>Bem-vindo ao MoneyTracker!</b>\n\nSeu assistente financeiro pessoal no Telegram.\nAcompanhe gastos, monitore receitas e gerencie seu dinheiro.\n\nToque no botão abaixo para abrir o app 👇",
		help:    "<b>Ajuda do MoneyTracker</b>\n\nAbra o Mini App para gerenciar suas finanças:\n• Adicione gastos e receitas\n• Veja saldo e histórico de transações\n• Acompanhe orçamentos e metas de poupança\n• Configure transações recorrentes\n• Exporte seus dados\n\nUse o botão abaixo para começar.",
		defaultMsg: "Use o botão abaixo para abrir o MoneyTracker ou digite /help para mais informações.",
		openButton: "📱 Abrir MoneyTracker",
	},
	domain.LangNL: {
		welcome: "<b>Welkom bij MoneyTracker!</b>\n\nJouw persoonlijke financiële assistent in Telegram.\nHoud uitgaven bij, monitor inkomsten en beheer je geld.\n\nTik op de knop hieronder om de app te openen 👇",
		help:    "<b>MoneyTracker Help</b>\n\nOpen de Mini App om je financiën te beheren:\n• Voeg uitgaven en inkomsten toe\n• Bekijk saldo en transactiegeschiedenis\n• Volg budgetten en spaardoelen\n• Stel terugkerende transacties in\n• Exporteer je gegevens\n\nGebruik de knop hieronder om te beginnen.",
		defaultMsg: "Gebruik de knop hieronder om MoneyTracker te openen of typ /help voor meer info.",
		openButton: "📱 MoneyTracker openen",
	},
	domain.LangAR: {
		welcome: "<b>مرحباً بك في MoneyTracker!</b>\n\nمساعدك المالي الشخصي على Telegram.\nتتبع المصروفات، وراقب الدخل، وأدر أموالك.\n\nاضغط على الزر أدناه لفتح التطبيق 👇",
		help:    "<b>مساعدة MoneyTracker</b>\n\nافتح التطبيق المصغر لإدارة أموالك:\n• أضف المصروفات والدخل\n• اعرض الرصيد وسجل المعاملات\n• تتبع الميزانيات وأهداف الادخار\n• أعد المعاملات المتكررة\n• صدّر بياناتك\n\nاستخدم الزر أدناه للبدء.",
		defaultMsg: "استخدم الزر أدناه لفتح MoneyTracker أو اكتب /help لمزيد من المعلومات.",
		openButton: "📱 فتح MoneyTracker",
	},
	domain.LangTR: {
		welcome: "<b>MoneyTracker'a Hoş Geldiniz!</b>\n\nTelegram'daki kişisel finans asistanınız.\nHarcamaları takip edin, gelirleri izleyin ve paranızı yönetin.\n\nUygulamayı açmak için aşağıdaki düğmeye dokunun 👇",
		help:    "<b>MoneyTracker Yardım</b>\n\nFinanslarınızı yönetmek için Mini Uygulamayı açın:\n• Gider ve gelir ekleyin\n• Bakiye ve işlem geçmişini görüntüleyin\n• Bütçeleri ve tasarruf hedeflerini takip edin\n• Tekrarlayan işlemleri ayarlayın\n• Verilerinizi dışa aktarın\n\nBaşlamak için aşağıdaki düğmeyi kullanın.",
		defaultMsg: "MoneyTracker'ı açmak için aşağıdaki düğmeyi kullanın veya daha fazla bilgi için /help yazın.",
		openButton: "📱 MoneyTracker'ı Aç",
	},
	domain.LangKO: {
		welcome: "<b>MoneyTracker에 오신 것을 환영합니다!</b>\n\nTelegram의 개인 금융 도우미입니다.\n지출을 추적하고, 수입을 모니터링하며, 재정을 관리하세요.\n\n아래 버튼을 눌러 앱을 여세요 👇",
		help:    "<b>MoneyTracker 도움말</b>\n\n재정 관리를 위해 미니 앱을 여세요:\n• 지출 및 수입 추가\n• 잔액 및 거래 내역 확인\n• 예산 및 저축 목표 추적\n• 반복 거래 설정\n• 데이터 내보내기\n\n시작하려면 아래 버튼을 사용하세요.",
		defaultMsg: "MoneyTracker를 열려면 아래 버튼을 사용하거나 /help를 입력하세요.",
		openButton: "📱 MoneyTracker 열기",
	},
	domain.LangMS: {
		welcome: "<b>Selamat datang ke MoneyTracker!</b>\n\nPembantu kewangan peribadi anda di Telegram.\nJejak perbelanjaan, pantau pendapatan, dan urus wang anda.\n\nKetik butang di bawah untuk membuka aplikasi 👇",
		help:    "<b>Bantuan MoneyTracker</b>\n\nBuka Mini App untuk mengurus kewangan anda:\n• Tambah perbelanjaan dan pendapatan\n• Lihat baki dan sejarah transaksi\n• Jejak belanjawan dan matlamat simpanan\n• Sediakan transaksi berulang\n• Eksport data anda\n\nGunakan butang di bawah untuk bermula.",
		defaultMsg: "Gunakan butang di bawah untuk membuka MoneyTracker atau taip /help untuk maklumat lanjut.",
		openButton: "📱 Buka MoneyTracker",
	},
	domain.LangID: {
		welcome: "<b>Selamat datang di MoneyTracker!</b>\n\nAsisten keuangan pribadi Anda di Telegram.\nLacak pengeluaran, pantau pendapatan, dan kelola keuangan Anda.\n\nKetuk tombol di bawah untuk membuka aplikasi 👇",
		help:    "<b>Bantuan MoneyTracker</b>\n\nBuka Mini App untuk mengelola keuangan Anda:\n• Tambahkan pengeluaran dan pendapatan\n• Lihat saldo dan riwayat transaksi\n• Lacak anggaran dan tujuan tabungan\n• Atur transaksi berulang\n• Ekspor data Anda\n\nGunakan tombol di bawah untuk memulai.",
		defaultMsg: "Gunakan tombol di bawah untuk membuka MoneyTracker atau ketik /help untuk info lebih lanjut.",
		openButton: "📱 Buka MoneyTracker",
	},
}

// getString returns localised bot strings for the given language,
// falling back to English if the language is not found.
func getString(lang domain.Language) botStrings {
	if s, ok := messages[lang]; ok {
		return s
	}
	return messages[domain.LangEN]
}

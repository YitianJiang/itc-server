package form

type CardElementForm struct {
	CardTextElementForm
	//--------------------Meta-----------------------
	Tag    string  `json:"tag,omitempty" validate:"omitempty"`
	Style  *string `json:"style,omitempty" validate:"omitempty"`
	UserId string  `json:"user_id,omitempty" validate:"omitempty"`
	//-------------------------------------------------

	//--------------------header-----------------------
	Title      *string `json:"title,omitempty" validate:"omitempty"`
	ImageColor string  `json:"image_color,omitempty" validate:"omitempty"`
	HideImage  bool    `json:"hide_image,omitempty" validate:"omitempty"`
	//-------------------------------------------------

	//--------------------Link-----------------------
	Href        string  `json:"href,omitempty" validate:"omitempty"`
	IOSHref     *string `json:"ios_href,omitempty" validate:"omitempty"`
	AndroidHref *string `json:"android_href,omitempty" validate:"omitempty"`
	PCHref      *string `json:"pc_href,omitempty" validate:"omitempty"`
	//-------------------------------------------------

	//--------------------Image-----------------------
	ImageKey string `json:"image_key,omitempty" validate:"omitempty"`
	Height   int32  `json:"height,omitempty" validate:"omitempty"`
	Width    int32  `json:"width,omitempty" validate:"omitempty"`
	//-------------------------------------------------

	//--------------------Devider-----------------------
	Orientation *string `json:"orientation,omitempty" validate:"omitempty"`
	Color       *string `json:"color,omitempty" validate:"omitempty"`
	Size        *string `json:"size,omitempty" validate:"omitempty"`
	//-------------------------------------------------

	//--------------------Fields-----------------------
	Fields []CardFieldForm `json:"fields,omitempty" validate:"omitempty"`
	//-------------------------------------------------

	//--------------------Sections-----------------------
	Sections []CardElementForm `json:"sections,omitempty" validate:"omitempty"`
	//-------------------------------------------------
}

type CardTextElementForm struct {
	Text          *string           `json:"text,omitempty" validate:"omitempty"`
	TextColor     string            `json:"text_color,omitempty" validate:"omitempty"`
	Lines         *int32            `json:"lines,omitempty" validate:"omitempty,min=1,max=100"`
	UnEscape      bool              `json:"un_escape,omitempty" validate:"omitempty"`
	I18n          map[string]string `json:"i18n,omitempty" validate:"omitempty"`
	TriggeredI18n map[string]string `json:"triggered_i18n,omitempty" validate:"omitempty"`
}

type CardFieldForm struct {
	Value *CardElementForm `json:"value,omitempty" validate:"omitempty"`
	Title *CardElementForm `json:"title,omitempty" validate:"omitempty"`
	Short *bool            `json:"short,omitempty" validate:"omitempty"`
}

type PostForm struct {
	Title   *string             `json:"title" validate:"omitempty"`
	Content [][]CardElementForm `json:"content" validate:"omitempty"`
}

type CardForm struct {
	CardLink *CardElementForm    `json:"card_link,omitempty" validate:"omitempty"`
	Header   *CardElementForm    `json:"header,omitempty" validate:"omitempty"`
	Content  [][]CardElementForm `json:"content,omitempty" validate:"omitempty"`
	Actions  []CardActionForm    `json:"actions,omitempty" validate:"omitempty"`
	ImageKeys     map[string]bool              `json:"-"`
	I18nArrayKeys map[string]map[string]string `json:"-"`
}

type CardActionForm struct {
	Changeable bool             `json:"changeable,omitempty" validate:"omitempty"`
	Buttons    []CardButtonForm `json:"buttons,omitempty" validate:"omitempty"`
}

type CardButtonForm struct {
	CardTextElementForm
	TriggeredText   *string                `json:"triggered_text,omitempty" validate:"omitempty"`
	HideOthers      *bool                  `json:"hide_others,omitempty" validate:"omitempty"`
	Url             *string                `json:"url,omitempty" validate:"omitempty"`
	Method          string                 `json:"method,omitempty" validate:"omitempty"`
	NeedUserInfo    *bool                  `json:"need_user_info,omitempty" validate:"omitempty"`
	NeedMessageInfo *bool                  `json:"need_message_info,omitempty" validate:"omitempty"`
	Parameter       map[string]interface{} `json:"parameter,omitempty" validate:"omitempty"`
	OpenUrl CardOpenUrlForm `json:"open_url,omitempty" validate:"omitempty"`
	Style   *string         `json:"style,omitempty" validate:"omitempty"`
}

type CardOpenUrlForm struct {
	IosUrl     *string `json:"ios_url,omitempty" validate:"omitempty"`
	PcUrl      *string `json:"pc_url,omitempty" validate:"omitempty"`
	AndroidUrl *string `json:"android_url,omitempty" validate:"omitempty"`
}
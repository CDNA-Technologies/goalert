package slack

import (
	"strconv"

	"github.com/slack-go/slack"
	"github.com/target/goalert/alert"
)

func alertIDAndStatusSection(alertID string, status alert.Status) *slack.HeaderBlock {
	var s string
	switch status {
	case alert.StatusTriggered:
		s = "Unacknowledged"
	case alert.StatusActive:
		s = "Acknowledged"
	case alert.StatusClosed:
		s = "Closed"
	default:
		panic("alert type not supported")
	}
	txt := alertID + ": " + s
	summaryText := slack.NewTextBlockObject("plain_text", txt, false, false)
	return slack.NewHeaderBlock(summaryText)
}

func alertSummarySection(summary string) *slack.SectionBlock {
	summaryText := slack.NewTextBlockObject("mrkdwn", "Summary: "+summary, false, false)
	return slack.NewSectionBlock(summaryText, nil, nil)
}

func ackButton(alertID string) slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Acknowledge :eyes:", true, false)
	return *slack.NewButtonBlockElement("ack", alertID, txt)
}

func escButton(alertID string) *slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Escalate :arrow_up:", true, false)
	return slack.NewButtonBlockElement("esc", alertID, txt)
}

func closeButton(alertID string) *slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Close :ballot_box_with_check:", true, false)
	return slack.NewButtonBlockElement("close", alertID, txt)
}

func openLinkButton(url string) *slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Open in GoAlert :link:", true, false)
	s := slack.NewButtonBlockElement("openLink", "", txt)
	s.URL = url
	return s
}

func alertLastStatusContext(lastStatus string) *slack.ContextBlock {
	lastStatusText := slack.NewTextBlockObject("plain_text", lastStatus, true, true)
	return slack.NewContextBlock("", []slack.MixedElement{lastStatusText}...)
}

func userAuthMessageOption(clientID, alertID, uri string) slack.MsgOption {
	msg := slack.NewTextBlockObject("plain_text", "Please link your GoAlert account to continue", false, false)

	btnTxt := slack.NewTextBlockObject("plain_text", "Authenticate :link:", true, false)
	btn := slack.NewButtonBlockElement("auth", alertID, btnTxt)
	btn.URL = "https://slack.com/oauth/v2/authorize?user_scope=identity.basic&client_id=" + clientID + "&redirect_uri=" + uri // slack oauth endpoint
	accessory := slack.NewAccessory(btn)

	section := slack.NewSectionBlock(msg, nil, accessory)
	return slack.MsgOptionBlocks(section)
}

func CraftAlertMessage(a alert.Alert, url string) []slack.MsgOption {
	var msgOpt []slack.MsgOption
	var actions *slack.ActionBlock

	alertID := strconv.Itoa(a.ID)

	if a.Status == alert.StatusTriggered {
		actions = slack.NewActionBlock("", ackButton(alertID), escButton(alertID), closeButton(alertID), openLinkButton(url))
	} else if a.Status == alert.StatusActive {
		actions = slack.NewActionBlock("", escButton(alertID), closeButton(alertID), openLinkButton(url))
	} else {
		actions = slack.NewActionBlock("", openLinkButton(url))
	}

	msgOpt = append(msgOpt,
		// desktop notification text
		slack.MsgOptionText(a.Summary, false),

		// blockkit elements
		slack.MsgOptionBlocks(
			alertIDAndStatusSection(alertID, a.Status),
			alertSummarySection(a.Summary),
			actions,
		),
	)

	return msgOpt
}
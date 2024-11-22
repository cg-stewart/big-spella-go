package chime

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chime"
	"github.com/aws/aws-sdk-go-v2/service/chime/types"
	"github.com/google/uuid"
)

type MeetingService struct {
	client *chime.Client
}

type MeetingInfo struct {
	MeetingID string
	ExternalMeetingID string
	MediaPlacement    *types.MediaPlacement
	Attendees        []AttendeeInfo
}

type AttendeeInfo struct {
	AttendeeID          string
	ExternalUserID      string
	JoinToken          string
}

func NewMeetingService(cfg aws.Config) *MeetingService {
	return &MeetingService{
		client: chime.NewFromConfig(cfg),
	}
}

// CreateGameMeeting creates a new Chime meeting for a game
func (s *MeetingService) CreateGameMeeting(ctx context.Context, gameID string) (*MeetingInfo, error) {
	// Create meeting
	meeting, err := s.client.CreateMeeting(ctx, &chime.CreateMeetingInput{
		ClientRequestToken: aws.String(uuid.New().String()),
		ExternalMeetingId: aws.String(fmt.Sprintf("game-%s", gameID)),
		MediaRegion:       aws.String("us-east-1"), // Configure based on game region
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create meeting: %w", err)
	}

	return &MeetingInfo{
		MeetingID:         aws.ToString(meeting.Meeting.MeetingId),
		ExternalMeetingID: aws.ToString(meeting.Meeting.ExternalMeetingId),
		MediaPlacement:    meeting.Meeting.MediaPlacement,
	}, nil
}

// AddAttendee adds a player to a meeting
func (s *MeetingService) AddAttendee(ctx context.Context, meetingID, userID string) (*AttendeeInfo, error) {
	attendee, err := s.client.CreateAttendee(ctx, &chime.CreateAttendeeInput{
		MeetingId:     aws.String(meetingID),
		ExternalUserId: aws.String(userID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create attendee: %w", err)
	}

	return &AttendeeInfo{
		AttendeeID:     aws.ToString(attendee.Attendee.AttendeeId),
		ExternalUserID: aws.ToString(attendee.Attendee.ExternalUserId),
		JoinToken:     aws.ToString(attendee.Attendee.JoinToken),
	}, nil
}

// DeleteMeeting ends a meeting
func (s *MeetingService) DeleteMeeting(ctx context.Context, meetingID string) error {
	_, err := s.client.DeleteMeeting(ctx, &chime.DeleteMeetingInput{
		MeetingId: aws.String(meetingID),
	})
	return err
}

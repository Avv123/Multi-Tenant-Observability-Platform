package kafka

import (
	"context"
	"sort"

	"github.com/IBM/sarama"
)

type ConsumerGroupLagRow struct {
	GroupID        string `json:"group_id"`
	Topic          string `json:"topic"`
	Partition      int32  `json:"partition"`
	CurrentOffset  int64  `json:"current_offset"`
	LatestOffset   int64  `json:"latest_offset"`
	Lag            int64  `json:"lag"`
	MemberAssigned bool   `json:"member_assigned"`
}

func Ping(_ context.Context, brokers []string) error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()
	return nil
}

func ConsumerGroupLag(brokers []string, groupID string, topics []string) ([]ConsumerGroupLagRow, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, err
	}
	defer func() { _ = client.Close() }()

	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		return nil, err
	}
	defer func() { _ = admin.Close() }()

	partitionsByTopic := make(map[string][]int32, len(topics))
	for _, topic := range topics {
		partitions, partitionsErr := client.Partitions(topic)
		if partitionsErr != nil {
			continue
		}
		partitionsByTopic[topic] = partitions
	}

	offsetFetch, err := admin.ListConsumerGroupOffsets(groupID, partitionsByTopic)
	if err != nil {
		return nil, err
	}

	rows := make([]ConsumerGroupLagRow, 0)
	for topic, partitions := range partitionsByTopic {
		for _, partition := range partitions {
			block := offsetFetch.GetBlock(topic, partition)
			if block == nil {
				continue
			}
			latestOffset, latestErr := client.GetOffset(topic, partition, sarama.OffsetNewest)
			if latestErr != nil {
				return nil, latestErr
			}
			currentOffset := block.Offset
			if currentOffset < 0 {
				currentOffset = 0
			}
			lag := latestOffset - currentOffset
			if lag < 0 {
				lag = 0
			}
			rows = append(rows, ConsumerGroupLagRow{
				GroupID:        groupID,
				Topic:          topic,
				Partition:      partition,
				CurrentOffset:  currentOffset,
				LatestOffset:   latestOffset,
				Lag:            lag,
				MemberAssigned: block.Offset >= 0,
			})
		}
	}

	sort.Slice(rows, func(left, right int) bool {
		if rows[left].Topic == rows[right].Topic {
			return rows[left].Partition < rows[right].Partition
		}
		return rows[left].Topic < rows[right].Topic
	})
	return rows, nil
}

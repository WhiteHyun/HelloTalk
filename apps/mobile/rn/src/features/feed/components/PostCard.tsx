import React from 'react';
import {View, Text, TouchableOpacity, StyleSheet, Alert} from 'react-native';
import {PostResponse} from '../../../infrastructure/api/feed';
import {useAuth} from '../../../app/AuthContext';

type Props = {
  post: PostResponse;
  onLikeToggle: () => void;
  onDelete: () => void;
};

export function PostCard({post, onLikeToggle, onDelete}: Props) {
  const {user} = useAuth();
  const isOwner = user?.id === post.user_id;

  const confirmDelete = () => {
    Alert.alert('게시글 삭제', '정말 삭제하시겠습니까?', [
      {text: '취소', style: 'cancel'},
      {text: '삭제', style: 'destructive', onPress: onDelete},
    ]);
  };

  const timeAgo = (dateStr: string) => {
    const diff = Date.now() - new Date(dateStr).getTime();
    const minutes = Math.floor(diff / 60000);
    if (minutes < 1) {
      return '방금 전';
    }
    if (minutes < 60) {
      return `${minutes}분 전`;
    }
    const hours = Math.floor(minutes / 60);
    if (hours < 24) {
      return `${hours}시간 전`;
    }
    const days = Math.floor(hours / 24);
    return `${days}일 전`;
  };

  return (
    <View style={styles.card}>
      <View style={styles.header}>
        <Text style={styles.userName}>{post.user_name}</Text>
        <Text style={styles.time}>{timeAgo(post.created_at)}</Text>
      </View>

      <Text style={styles.content}>{post.content}</Text>

      <View style={styles.footer}>
        <TouchableOpacity onPress={onLikeToggle} style={styles.likeButton}>
          <Text style={[styles.likeText, post.liked && styles.liked]}>
            {post.liked ? '♥' : '♡'} {post.like_count}
          </Text>
        </TouchableOpacity>

        {isOwner && (
          <TouchableOpacity onPress={confirmDelete}>
            <Text style={styles.deleteText}>삭제</Text>
          </TouchableOpacity>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#fff',
    marginHorizontal: 16,
    marginTop: 12,
    borderRadius: 12,
    padding: 16,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  userName: {
    fontSize: 15,
    fontWeight: '600',
  },
  time: {
    fontSize: 12,
    color: '#999',
  },
  content: {
    fontSize: 15,
    lineHeight: 22,
    color: '#333',
    marginBottom: 12,
  },
  footer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  likeButton: {
    paddingVertical: 4,
  },
  likeText: {
    fontSize: 14,
    color: '#666',
  },
  liked: {
    color: '#FF3B30',
  },
  deleteText: {
    fontSize: 13,
    color: '#FF3B30',
  },
});

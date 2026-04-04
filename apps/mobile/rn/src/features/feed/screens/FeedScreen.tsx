import React, {useCallback, useEffect, useState} from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  RefreshControl,
  Alert,
} from 'react-native';
import {NativeStackNavigationProp} from '@react-navigation/native-stack';
import * as feedApi from '../../../infrastructure/api/feed';
import {PostResponse} from '../../../infrastructure/api/feed';
import {PostCard} from '../components/PostCard';

type Props = {
  navigation: NativeStackNavigationProp<any>;
};

export function FeedScreen({navigation}: Props) {
  const [posts, setPosts] = useState<PostResponse[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);

  const loadFeed = useCallback(async (cursor?: string) => {
    try {
      const result = await feedApi.getFeed(cursor);
      if (cursor) {
        setPosts(prev => [...prev, ...result.posts]);
      } else {
        setPosts(result.posts);
      }
      setNextCursor(result.next_cursor);
    } catch {
      Alert.alert('오류', '피드를 불러올 수 없습니다');
    }
  }, []);

  useEffect(() => {
    loadFeed();
  }, [loadFeed]);

  const onRefresh = async () => {
    setRefreshing(true);
    await loadFeed();
    setRefreshing(false);
  };

  const onEndReached = async () => {
    if (!nextCursor || loadingMore) {
      return;
    }
    setLoadingMore(true);
    await loadFeed(nextCursor);
    setLoadingMore(false);
  };

  const handleLikeToggle = async (post: PostResponse) => {
    try {
      if (post.liked) {
        await feedApi.unlikePost(post.id);
      } else {
        await feedApi.likePost(post.id);
      }
      setPosts(prev =>
        prev.map(p =>
          p.id === post.id
            ? {
                ...p,
                liked: !p.liked,
                like_count: p.liked ? p.like_count - 1 : p.like_count + 1,
              }
            : p,
        ),
      );
    } catch {
      Alert.alert('오류', '좋아요 처리에 실패했습니다');
    }
  };

  const handleDelete = async (postId: string) => {
    try {
      await feedApi.deletePost(postId);
      setPosts(prev => prev.filter(p => p.id !== postId));
    } catch {
      Alert.alert('오류', '삭제에 실패했습니다');
    }
  };

  return (
    <View style={styles.container}>
      <FlatList
        data={posts}
        keyExtractor={item => item.id}
        renderItem={({item}) => (
          <PostCard
            post={item}
            onLikeToggle={() => handleLikeToggle(item)}
            onDelete={() => handleDelete(item.id)}
          />
        )}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        onEndReached={onEndReached}
        onEndReachedThreshold={0.5}
        ListEmptyComponent={
          <Text style={styles.empty}>아직 게시글이 없습니다</Text>
        }
      />
      <TouchableOpacity
        style={styles.fab}
        onPress={() => navigation.navigate('CreatePost')}>
        <Text style={styles.fabText}>+</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  empty: {
    textAlign: 'center',
    color: '#999',
    marginTop: 48,
    fontSize: 16,
  },
  fab: {
    position: 'absolute',
    right: 20,
    bottom: 20,
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: '#007AFF',
    justifyContent: 'center',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.25,
    shadowRadius: 4,
    elevation: 5,
  },
  fabText: {
    color: '#fff',
    fontSize: 28,
    fontWeight: '300',
    marginTop: -2,
  },
});

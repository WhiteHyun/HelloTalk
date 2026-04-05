import React, {useState} from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  Alert,
  ActivityIndicator,
} from 'react-native';
import {NativeStackNavigationProp} from '@react-navigation/native-stack';
import * as feedApi from '../../../infrastructure/api/feed';

type Props = {
  navigation: NativeStackNavigationProp<any>;
};

export function CreatePostScreen({navigation}: Props) {
  const [content, setContent] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    if (!content.trim()) {
      Alert.alert('오류', '내용을 입력해주세요');
      return;
    }

    setLoading(true);
    try {
      await feedApi.createPost(content.trim());
      navigation.goBack();
    } catch {
      Alert.alert('오류', '게시글 작성에 실패했습니다');
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.container}>
      <TextInput
        style={styles.input}
        placeholder="무슨 생각을 하고 있나요?"
        value={content}
        onChangeText={setContent}
        multiline
        autoFocus
        textAlignVertical="top"
      />
      <TouchableOpacity
        style={[styles.button, !content.trim() && styles.buttonDisabled]}
        onPress={handleSubmit}
        disabled={loading || !content.trim()}>
        {loading ? (
          <ActivityIndicator color="#fff" />
        ) : (
          <Text style={styles.buttonText}>게시</Text>
        )}
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
    backgroundColor: '#fff',
  },
  input: {
    flex: 1,
    fontSize: 16,
    lineHeight: 24,
    padding: 0,
    marginBottom: 16,
  },
  button: {
    height: 48,
    backgroundColor: '#007AFF',
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
  },
  buttonDisabled: {
    opacity: 0.5,
  },
  buttonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
});

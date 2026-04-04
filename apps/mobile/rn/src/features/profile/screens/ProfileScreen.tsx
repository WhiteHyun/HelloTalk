import React, {useState} from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  Alert,
} from 'react-native';
import {useAuth} from '../../../app/AuthContext';
import * as userApi from '../../../infrastructure/api/user';

export function ProfileScreen() {
  const {user, signOut} = useAuth();
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(user?.name ?? '');

  const handleSave = async () => {
    if (!name.trim()) {
      Alert.alert('오류', '이름을 입력해주세요');
      return;
    }
    try {
      await userApi.updateMe(name.trim());
      setEditing(false);
      Alert.alert('완료', '프로필이 수정되었습니다');
    } catch {
      Alert.alert('오류', '프로필 수정에 실패했습니다');
    }
  };

  const handleLogout = () => {
    Alert.alert('로그아웃', '정말 로그아웃 하시겠습니까?', [
      {text: '취소', style: 'cancel'},
      {text: '로그아웃', style: 'destructive', onPress: signOut},
    ]);
  };

  return (
    <View style={styles.container}>
      <View style={styles.avatarContainer}>
        <View style={styles.avatar}>
          <Text style={styles.avatarText}>
            {user?.name?.charAt(0)?.toUpperCase() ?? '?'}
          </Text>
        </View>
      </View>

      <View style={styles.infoSection}>
        <Text style={styles.label}>이메일</Text>
        <Text style={styles.value}>{user?.email}</Text>
      </View>

      <View style={styles.infoSection}>
        <Text style={styles.label}>이름</Text>
        {editing ? (
          <View style={styles.editRow}>
            <TextInput
              style={styles.editInput}
              value={name}
              onChangeText={setName}
              autoFocus
            />
            <TouchableOpacity onPress={handleSave}>
              <Text style={styles.saveButton}>저장</Text>
            </TouchableOpacity>
            <TouchableOpacity onPress={() => setEditing(false)}>
              <Text style={styles.cancelButton}>취소</Text>
            </TouchableOpacity>
          </View>
        ) : (
          <TouchableOpacity onPress={() => setEditing(true)}>
            <Text style={styles.value}>
              {user?.name} <Text style={styles.editHint}>수정</Text>
            </Text>
          </TouchableOpacity>
        )}
      </View>

      <View style={styles.infoSection}>
        <Text style={styles.label}>가입일</Text>
        <Text style={styles.value}>
          {user?.created_at
            ? new Date(user.created_at).toLocaleDateString('ko-KR')
            : '-'}
        </Text>
      </View>

      <TouchableOpacity style={styles.logoutButton} onPress={handleLogout}>
        <Text style={styles.logoutText}>로그아웃</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
    padding: 24,
  },
  avatarContainer: {
    alignItems: 'center',
    marginBottom: 32,
    marginTop: 16,
  },
  avatar: {
    width: 80,
    height: 80,
    borderRadius: 40,
    backgroundColor: '#007AFF',
    justifyContent: 'center',
    alignItems: 'center',
  },
  avatarText: {
    color: '#fff',
    fontSize: 32,
    fontWeight: 'bold',
  },
  infoSection: {
    marginBottom: 24,
  },
  label: {
    fontSize: 13,
    color: '#999',
    marginBottom: 4,
  },
  value: {
    fontSize: 16,
    color: '#333',
  },
  editHint: {
    fontSize: 13,
    color: '#007AFF',
  },
  editRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  editInput: {
    flex: 1,
    height: 40,
    borderWidth: 1,
    borderColor: '#ddd',
    borderRadius: 8,
    paddingHorizontal: 12,
    fontSize: 16,
  },
  saveButton: {
    color: '#007AFF',
    fontSize: 15,
    fontWeight: '600',
  },
  cancelButton: {
    color: '#999',
    fontSize: 15,
  },
  logoutButton: {
    marginTop: 'auto',
    height: 48,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#FF3B30',
    justifyContent: 'center',
    alignItems: 'center',
  },
  logoutText: {
    color: '#FF3B30',
    fontSize: 16,
    fontWeight: '600',
  },
});

export interface TrainingRegistration {
  attendance_id: number;
  training_id: number;
  date: string;
  time: string;
  end_time: string;
  group_name: string;
  coach_name: string;
  description: string;
  status: string;
  attended: boolean;
  notes?: string;
}

export interface MyRegistrationsResponse {
  upcoming: TrainingRegistration[];
  past: TrainingRegistration[];
}

export interface Subscription {
  id: number;
  total_lessons: number;
  remaining_lessons: number;
  used_lessons: number;
  start_date: string;
  end_date: string;
  created_at: string;
  status?: string; // для истекших абонементов: "used" или "expired"
}

export interface MySubscriptionResponse {
  active: Subscription[];
  expired: Subscription[];
  has_active: boolean;
}

export interface UserProfile {
  id: number;
  first_name: string;
  last_name: string;
  username: string;
  role: string;
  registered_at: string;
}

export interface StudentProfile {
  id: number;
  athletic_title: string;
  created_at: string;
}

export interface MyProfileResponse {
  user: UserProfile;
  student: StudentProfile;
}

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
  student?: StudentProfile;
  coach?: CoachProfile;
}

export interface CoachProfile {
  id: number;
  specialty: string;
  experience: string;
  description: string;
  created_at: string;
}

export interface StudentWithSubscriptions {
  user_id: number;
  student_id: number;
  name: string;
  subscriptions: Subscription[];
}

export interface AllStudentsSubscriptionsResponse {
  students: StudentWithSubscriptions[];
}

export interface UpdateProfileRequest {
  first_name?: string;
  last_name?: string;
  athletic_title?: string; // for students
  specialty?: string; // for coaches
  experience?: string; // for coaches
  description?: string; // for coaches
}
